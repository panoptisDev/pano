// Copyright 2025 Pano Operations Ltd
// This file is part of the Pano Client
//
// Pano is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Pano is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Pano. If not, see <http://www.gnu.org/licenses/>.

package gossip

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand/v2"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/dag"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/lachesis"
	"github.com/panoptisDev/lachesis-base/utils/workers"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/dnsdisc"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/panoptisDev/pano/ethapi"
	"github.com/panoptisDev/pano/eventcheck"
	"github.com/panoptisDev/pano/eventcheck/basiccheck"
	"github.com/panoptisDev/pano/eventcheck/epochcheck"
	"github.com/panoptisDev/pano/eventcheck/gaspowercheck"
	"github.com/panoptisDev/pano/eventcheck/heavycheck"
	"github.com/panoptisDev/pano/eventcheck/parentscheck"
	"github.com/panoptisDev/pano/eventcheck/proposalcheck"
	"github.com/panoptisDev/pano/evmcore"
	"github.com/panoptisDev/pano/gossip/blockproc"
	"github.com/panoptisDev/pano/gossip/blockproc/drivermodule"
	"github.com/panoptisDev/pano/gossip/blockproc/eventmodule"
	"github.com/panoptisDev/pano/gossip/blockproc/evmmodule"
	"github.com/panoptisDev/pano/gossip/blockproc/sealmodule"
	"github.com/panoptisDev/pano/gossip/blockproc/verwatcher"
	"github.com/panoptisDev/pano/gossip/emitter"
	"github.com/panoptisDev/pano/gossip/evmstore"
	"github.com/panoptisDev/pano/gossip/filters"
	"github.com/panoptisDev/pano/gossip/gasprice"
	"github.com/panoptisDev/pano/gossip/proclogger"
	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/logger"
	scc_node "github.com/panoptisDev/pano/scc/node"
	"github.com/panoptisDev/pano/utils/signers/gsignercache"
	"github.com/panoptisDev/pano/utils/txtime"
	"github.com/panoptisDev/pano/utils/wgmutex"
	"github.com/panoptisDev/pano/valkeystore"
	"github.com/panoptisDev/pano/vecmt"
)

//go:generate mockgen -source=service.go -package=gossip -destination=service_mock.go

type ServiceFeed struct {
	scope notify.SubscriptionScope

	newEpoch        notify.Feed
	newEmittedEvent notify.Feed
	newBlock        notify.Feed
	newLogs         notify.Feed

	incomingUpdates chan<- feedUpdate // < channel to send updates to the background feed loop
	stopFeeder      chan<- struct{}   // < if closed, the background feed loop will stop
	feederDone      <-chan struct{}   // < if closed, the background feed loop has stopped
}

type feedUpdate struct {
	block *evmcore.EvmBlock
	logs  []*types.Log
}

type ArchiveBlockHeightSource interface {
	GetArchiveBlockHeight() (uint64, bool, error)
}

func (f *ServiceFeed) SubscribeNewEpoch(ch chan<- idx.Epoch) notify.Subscription {
	return f.scope.Track(f.newEpoch.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewEmitted(ch chan<- *inter.EventPayload) notify.Subscription {
	return f.scope.Track(f.newEmittedEvent.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewBlock(ch chan<- evmcore.ChainHeadNotify) notify.Subscription {
	return f.scope.Track(f.newBlock.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewLogs(ch chan<- []*types.Log) notify.Subscription {
	return f.scope.Track(f.newLogs.Subscribe(ch))
}

func (f *ServiceFeed) Start(store ArchiveBlockHeightSource) {
	incoming := make(chan feedUpdate, 1024)
	f.incomingUpdates = incoming
	stop := make(chan struct{})
	done := make(chan struct{})
	f.stopFeeder = stop
	f.feederDone = done
	go func() {
		defer close(done)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		pending := []feedUpdate{}
		for {
			select {
			case <-stop:
				return
			case update := <-incoming:
				pending = append(pending, update)
				// sorting could be replaced by a heap or skipped if updates
				// are guaranteed to be delivered in order.
				slices.SortFunc(pending, func(a, b feedUpdate) int {
					return a.block.Number.Cmp(b.block.Number)
				})

			case <-ticker.C:
			}

			if len(pending) == 0 {
				continue
			}

			height, empty, err := store.GetArchiveBlockHeight()
			if err != nil {
				// If there is no archive, set height to the last block
				// and send all notifications
				if errors.Is(err, evmstore.NoArchiveError) {
					height = pending[len(pending)-1].block.Number.Uint64()
				} else {
					log.Error("failed to get archive block height", "err", err)
					continue
				}
			} else {
				if empty {
					continue
				}
			}

			for _, update := range pending {
				if update.block.Number.Uint64() > height {
					break
				}
				f.newBlock.Send(evmcore.ChainHeadNotify{Block: update.block})
				f.newLogs.Send(update.logs)
				pending = pending[1:]
			}
		}
	}()
}

func (f *ServiceFeed) notifyAboutNewBlock(
	block *evmcore.EvmBlock,
	logs []*types.Log,
) {
	f.incomingUpdates <- feedUpdate{
		block: block,
		logs:  logs,
	}
}

func (f *ServiceFeed) Stop() {
	if f.stopFeeder == nil {
		return
	}
	close(f.stopFeeder)
	f.stopFeeder = nil
	<-f.feederDone
	f.scope.Close()
}

type BlockProc struct {
	SealerModule     blockproc.SealerModule
	TxListenerModule blockproc.TxListenerModule
	PreTxTransactor  blockproc.TxTransactor
	PostTxTransactor blockproc.TxTransactor
	EventsModule     blockproc.ConfirmedEventsModule
	EVMModule        blockproc.EVM
}

func DefaultBlockProc() BlockProc {
	return BlockProc{
		SealerModule:     sealmodule.New(),
		TxListenerModule: drivermodule.NewDriverTxListenerModule(),
		PreTxTransactor:  drivermodule.NewDriverTxPreTransactor(),
		PostTxTransactor: drivermodule.NewDriverTxTransactor(),
		EventsModule:     eventmodule.New(),
		EVMModule:        evmmodule.New(),
	}
}

// Service implements go-ethereum/node.Service interface.
type Service struct {
	config Config

	// server
	p2pServer *p2p.Server
	Name      string

	accountManager *accounts.Manager

	// application
	store               *Store
	engine              lachesis.Consensus
	dagIndexer          *vecmt.Index
	engineMu            *sync.RWMutex
	emitters            []*emitter.Emitter
	txpool              TxPool
	heavyCheckReader    HeavyCheckReader
	gasPowerCheckReader GasPowerCheckReader
	proposalCheckReader proposalCheckReader
	checkers            *eventcheck.Checkers
	uniqueEventIDs      uniqueID

	// version watcher
	verWatcher *verwatcher.VersionWatcher

	// SCC node
	sccNode *scc_node.Node

	blockProcWg        sync.WaitGroup
	blockProcTasks     *workers.Workers
	blockProcTasksDone chan struct{}
	blockProcModules   BlockProc

	blockBusyFlag uint32
	eventBusyFlag uint32

	feed ServiceFeed

	gpo *gasprice.Oracle

	// application protocol
	handler *handler

	operaDialCandidates enode.Iterator

	EthAPI        *EthAPIBackend
	netRPCService *ethapi.PublicNetAPI

	procLogger *proclogger.Logger

	stopped   bool
	haltCheck func(oldEpoch, newEpoch idx.Epoch, time time.Time) bool

	tflusher PeriodicFlusher

	bootstrapping bool

	logger.Instance
}

func NewService(stack *node.Node, config Config, store *Store, blockProc BlockProc,
	engine lachesis.Consensus, dagIndexer *vecmt.Index, newTxPool func(evmcore.StateReader) TxPool,
	haltCheck func(oldEpoch, newEpoch idx.Epoch, age time.Time) bool) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	localNodeId := enode.PubkeyToIDV4(&stack.Server().PrivateKey.PublicKey)
	svc, err := newService(config, store, blockProc, engine, dagIndexer, newTxPool, localNodeId)
	if err != nil {
		return nil, err
	}

	svc.p2pServer = stack.Server()
	svc.accountManager = stack.AccountManager()
	svc.EthAPI.SetExtRPCEnabled(stack.Config().ExtRPCEnabled())
	// Create the net API service
	svc.netRPCService = ethapi.NewPublicNetAPI(svc.p2pServer, store.GetRules().NetworkID)
	svc.haltCheck = haltCheck

	return svc, nil
}

func newService(config Config, store *Store, blockProc BlockProc, engine lachesis.Consensus, dagIndexer *vecmt.Index, newTxPool func(evmcore.StateReader) TxPool, localId enode.ID) (*Service, error) {
	svc := &Service{
		config:             config,
		blockProcTasksDone: make(chan struct{}),
		Name:               fmt.Sprintf("Node-%d", rand.Int()),
		store:              store,
		engine:             engine,
		blockProcModules:   blockProc,
		dagIndexer:         dagIndexer,
		engineMu:           new(sync.RWMutex),
		uniqueEventIDs:     uniqueID{new(big.Int)},
		procLogger:         proclogger.NewLogger(),
		Instance:           logger.New("gossip-service"),
	}

	svc.blockProcTasks = workers.New(new(sync.WaitGroup), svc.blockProcTasksDone, 1)

	// load epoch DB
	svc.store.loadEpochStore(svc.store.GetEpoch())
	es := svc.store.getEpochStore(svc.store.GetEpoch())
	svc.dagIndexer.Reset(svc.store.GetValidators(), es.table.DagIndex, func(id hash.Event) dag.Event {
		return svc.store.GetEvent(id)
	})

	// load caches for mutable values to avoid race condition
	svc.store.GetBlockEpochState()
	svc.store.GetHighestLamport()
	svc.store.GetUpgradeHeights()
	svc.store.GetGenesisID()
	netVerStore := verwatcher.NewStore(store.table.NetworkVersion)
	netVerStore.GetNetworkVersion()
	netVerStore.GetMissedVersion()

	// create checkers
	net := store.GetRules()
	txSigner := gsignercache.Wrap(types.LatestSignerForChainID(new(big.Int).SetUint64(net.NetworkID)))
	svc.heavyCheckReader.Store = store
	svc.heavyCheckReader.Pubkeys.Store(readEpochPubKeys(svc.store, svc.store.GetEpoch()))                                          // read pub keys of current epoch from DB
	svc.gasPowerCheckReader.Ctx.Store(NewGasPowerContext(svc.store, svc.store.GetValidators(), svc.store.GetEpoch(), net.Economy)) // read gaspower check data from DB
	svc.proposalCheckReader = newProposalCheckReader(store)
	svc.checkers = makeCheckers(config.HeavyCheck, txSigner, &svc.heavyCheckReader, &svc.gasPowerCheckReader, &svc.proposalCheckReader, svc.store)

	// create GPO
	svc.gpo = gasprice.NewOracle(svc.config.GPO, nil)

	// create tx pool
	stateReader := &EvmStateReader{
		ServiceFeed: &svc.feed,
		store:       svc.store,
		gpo:         svc.gpo,
	}
	svc.txpool = newTxPool(stateReader)
	svc.gpo.SetReader(&GPOBackend{svc.store, svc.txpool})

	// init dialCandidates
	dnsclient := dnsdisc.NewClient(dnsdisc.Config{})
	var err error
	svc.operaDialCandidates, err = dnsclient.NewIterator(config.OperaDiscoveryURLs...)
	if err != nil {
		return nil, err
	}

	// create protocol manager
	svc.handler, err = newHandler(handlerConfig{
		config:   config,
		notifier: &svc.feed,
		txpool:   svc.txpool,
		engineMu: svc.engineMu,
		checkers: svc.checkers,
		s:        store,
		localId:  localId,
		process: processCallback{
			Event: func(event *inter.EventPayload) error {
				done := svc.procLogger.EventConnectionStarted(event, false)
				defer done()
				return svc.processEvent(event)
			},
			SwitchEpochTo: svc.SwitchEpochTo,
		},
		localEndPointSource: localEndPointSource{svc},
	})
	if err != nil {
		return nil, err
	}

	rpc.SetExecutionTimeLimit(config.RPCTimeout)

	// create API backend
	svc.EthAPI = &EthAPIBackend{false, svc, stateReader, txSigner, config.AllowUnprotectedTxs}

	svc.verWatcher = verwatcher.New(netVerStore)
	svc.tflusher = svc.makePeriodicFlusher()

	// create Pano Certification Chain node
	// TODO: track the current committee inside the scc Node instance
	// (see https://github.com/panoptisDev/pano-admin/issues/22)
	genesisCommitteeCertificate, err := store.GetCommitteeCertificate(0)
	if err == nil {
		svc.sccNode = scc_node.NewNode(store, genesisCommitteeCertificate.Subject().Committee)
	}

	return svc, nil
}

type localEndPointSource struct {
	service *Service
}

func (s localEndPointSource) GetLocalEndPoint() *enode.Node {
	return s.service.p2pServer.LocalNode().Node()
}

// makeCheckers builds event checkers
func makeCheckers(heavyCheckCfg heavycheck.Config, txSigner types.Signer, heavyCheckReader *HeavyCheckReader, gasPowerCheckReader *GasPowerCheckReader, proposalCheckReader *proposalCheckReader, store *Store) *eventcheck.Checkers {
	// create signatures checker
	heavyCheck := heavycheck.New(heavyCheckCfg, heavyCheckReader, txSigner)

	// create gaspower checker
	gaspowerCheck := gaspowercheck.New(gasPowerCheckReader)

	// create proposal checker
	proposalCheck := proposalcheck.New(proposalCheckReader)

	return &eventcheck.Checkers{
		Basiccheck:    basiccheck.New(),
		Epochcheck:    epochcheck.New(store),
		Parentscheck:  parentscheck.New(),
		Proposalcheck: proposalCheck,
		Heavycheck:    heavyCheck,
		Gaspowercheck: gaspowerCheck,
	}
}

// makePeriodicFlusher makes PeriodicFlusher
func (s *Service) makePeriodicFlusher() PeriodicFlusher {
	// Normally the diffs are committed by message processing. Yet, since we have async EVM snapshots generation,
	// we need to periodically commit its data
	return PeriodicFlusher{
		period: 10 * time.Millisecond,
		callback: PeriodicFlusherCallaback{
			busy: func() bool {
				// try to lock engineMu/blockProcWg pair as rarely as possible to not hurt
				// events/blocks pipeline concurrency
				return atomic.LoadUint32(&s.eventBusyFlag) != 0 || atomic.LoadUint32(&s.blockBusyFlag) != 0
			},
			commitNeeded: func() bool {
				// use slightly higher size threshold to avoid locking the mutex/wg pair and hurting events/blocks concurrency
				// PeriodicFlusher should mostly commit only data generated by async EVM snapshots generation
				return s.store.isCommitNeeded(1200, 1000)
			},
			commit: func() {
				s.engineMu.Lock()
				defer s.engineMu.Unlock()
				// Note: blockProcWg.Wait() is already called by s.commit
				if s.store.isCommitNeeded(1200, 1000) {
					s.commit(false)
				}
			},
		},
		wg:   sync.WaitGroup{},
		quit: make(chan struct{}),
	}
}

func (s *Service) EmitterWorld(signer valkeystore.SignerAuthority) emitter.World {
	return emitter.World{
		External: &emitterWorld{
			emitterWorldProc: emitterWorldProc{s},
			emitterWorldRead: emitterWorldRead{s.store},
			WgMutex:          wgmutex.New(s.engineMu, &s.blockProcWg),
		},
		TxPool:            s.txpool,
		EventsSigner:      signer,
		TransactionSigner: s.EthAPI.signer,
	}
}

// RegisterEmitter must be called before service is started
func (s *Service) RegisterEmitter(em *emitter.Emitter) {
	txtime.Enabled.Store(true) // enable tracking of tx times
	s.emitters = append(s.emitters, em)
}

type CleanupFunc func()

// MakeProtocols constructs the P2P protocol definitions for `opera`.
func MakeProtocols(svc *Service, backend *handler, disc enode.Iterator) ([]p2p.Protocol, CleanupFunc) {
	nodeIter := enode.NewFairMix(time.Second)
	nodeIter.AddSource(disc)
	nodeIter.AddSource(backend.GetSuggestedPeerIterator())

	protocols := make([]p2p.Protocol, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		version := version // Closure

		protocols[i] = p2p.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  protocolLengths[version],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				// wait until handler has started
				backend.started.Wait()
				peer := newPeer(version, p, rw, backend.config.Protocol.PeerCache)
				defer peer.Close()

				select {
				case <-backend.quitSync:
					return p2p.DiscQuitting
				default:
					backend.wg.Add(1)
					defer backend.wg.Done()
					return backend.handle(peer)
				}
			},
			NodeInfo: func() interface{} {
				return backend.NodeInfo()
			},
			PeerInfo: func(id enode.ID) interface{} {
				if p := backend.peers.Peer(id.String()); p != nil {
					return p.Info()
				}
				return nil
			},
			Attributes: []enr.Entry{
				currentENREntry(svc,
					0, // block height
					0, // time
				)},
			DialCandidates: nodeIter,
		}
	}
	return protocols, CleanupFunc(nodeIter.Close)
}

// Protocols returns protocols the service can communicate on.
func (s *Service) Protocols() ([]p2p.Protocol, CleanupFunc) {
	return MakeProtocols(s, s.handler, s.operaDialCandidates)
}

// APIs returns api methods the service wants to expose on rpc channels.
func (s *Service) APIs() []rpc.API {
	apis := ethapi.GetAPIs(s.EthAPI)

	apis = append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.EthAPI, s.config.FilterAPI),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   ethapi.NewPublicDebugAPI(s.EthAPI, s.config.MaxResponseSize, s.config.StructLogLimit),
			Public:    true,
		}, {
			Namespace: "trace",
			Version:   "1.0",
			Service:   ethapi.NewPublicTxTraceAPI(s.EthAPI, s.config.MaxResponseSize),
			Public:    true,
		}, {
			Namespace: "pano",
			Version:   "1.0",
			Service:   ethapi.NewPublicSccApi(s.EthAPI),
			Public:    true,
		},
	}...)

	// eth-namespace is doubled as ftm-namespace for branding purpose
	for _, api := range apis {
		if api.Namespace == "eth" {
			apis = append(apis, rpc.API{
				Namespace: "ftm",
				Version:   api.Version,
				Service:   api.Service,
				Public:    api.Public,
			})
		}
	}

	return apis
}

// Start method invoked when the node is ready to start the service.
func (s *Service) Start() error {
	s.gpo.Start()
	// start tflusher before starting snapshots generation
	s.tflusher.Start()
	blockState := s.store.GetBlockState()
	if s.store.evm.CheckLiveStateHash(blockState.LastBlock.Idx, blockState.FinalizedStateRoot) != nil {
		return errors.New("fullsync isn't possible because state root is missing")
	}

	// start notification feeder
	s.feed.Start(s.store.evm)

	// start blocks processor
	s.blockProcTasks.Start(1)

	// start p2p
	StartENRUpdater(s, s.p2pServer.LocalNode())
	s.handler.Start(s.p2pServer.MaxPeers)

	// start emitters
	for _, em := range s.emitters {
		em.Start()
	}

	s.verWatcher.Start()

	if s.haltCheck != nil && s.haltCheck(s.store.GetEpoch(), s.store.GetEpoch(), s.store.GetBlockState().LastBlock.Time.Time()) {
		// halt syncing
		s.stopped = true
	}

	return nil
}

// WaitBlockEnd waits until parallel block processing is complete (if any)
func (s *Service) WaitBlockEnd() {
	s.blockProcWg.Wait()
}

// Stop method invoked when the node terminates the service.
func (s *Service) Stop() error {
	defer log.Info("Pano service stopped")

	s.txpool.Stop()

	s.verWatcher.Stop()
	for _, em := range s.emitters {
		em.Stop()
	}

	// Stop all the peer-related stuff first.
	s.operaDialCandidates.Close()

	s.handler.Stop()
	s.feed.Stop()
	s.gpo.Stop()
	// it's safe to stop tflusher only before locking engineMu
	s.tflusher.Stop()

	// flush the state at exit, after all the routines stopped
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	s.stopped = true

	s.blockProcWg.Wait()
	close(s.blockProcTasksDone)

	err := s.dagIndexer.Close()
	if err != nil {
		return err
	}

	return s.store.Commit()
}

// AccountManager return node's account manager
func (s *Service) AccountManager() *accounts.Manager {
	return s.accountManager
}

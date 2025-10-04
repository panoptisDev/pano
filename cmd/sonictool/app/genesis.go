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

package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/panoptisDev/pano/cmd/panotool/db"
	"github.com/panoptisDev/pano/cmd/panotool/genesis"
	"github.com/panoptisDev/pano/config/flags"
	"github.com/panoptisDev/pano/integration/makefakegenesis"
	"github.com/panoptisDev/pano/opera"
	"github.com/panoptisDev/pano/opera/genesisstore"
	futils "github.com/panoptisDev/pano/utils"
	"github.com/panoptisDev/pano/utils/caution"
	"github.com/panoptisDev/pano/utils/memory"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/log"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"gopkg.in/urfave/cli.v1"
)

var (
	ModeFlag = cli.StringFlag{
		Name:  "mode",
		Usage: `Mode of the node ("rpc" or "validator")`,
		Value: "rpc",
	}
	ExperimentalFlag = cli.BoolFlag{
		Name:  "experimental",
		Usage: "Allow experimental features",
	}
	FakeUpgrades = cli.StringFlag{
		Name:  "upgrades",
		Usage: "Feature set enabled in the fake network, pano|allegro.",
		Value: "pano",
	}
)

func gfileGenesisImport(ctx *cli.Context) (err error) {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("this command requires an argument - the genesis file to import")
	}
	dataDir := ctx.GlobalString(flags.DataDirFlag.Name)
	if dataDir == "" {
		return fmt.Errorf("--%s need to be set", flags.DataDirFlag.Name)
	}
	validatorMode, err := isValidatorModeSet(ctx)
	if err != nil {
		return err
	}
	cacheRatio, err := cacheScaler(ctx)
	if err != nil {
		return err
	}

	genesisReader, err := os.Open(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("failed to open the genesis file: %w", err)
	}
	// note, genesisStore closes the reader, no need to defer close it here

	genesisStore, genesisHashes, err := genesisstore.OpenGenesisStore(genesisReader)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to read genesis file: %w", err), genesisReader.Close())
	}
	defer caution.CloseAndReportError(&err, genesisStore, "failed to close the genesis store")
	if err := genesis.IsGenesisTrusted(genesisStore, genesisHashes); err != nil {
		if ctx.IsSet(ExperimentalFlag.Name) {
			log.Warn("Experimental genesis file is used", "err", err)
		} else {
			return fmt.Errorf("genesis file check failed: %w", err)
		}
	}
	return genesis.ImportGenesisStore(genesis.ImportParams{
		GenesisStore:  genesisStore,
		DataDir:       dataDir,
		ValidatorMode: validatorMode,
		CacheRatio:    cacheRatio,
		LiveDbCache:   ctx.GlobalInt64(flags.LiveDbCacheFlag.Name),
		ArchiveCache:  ctx.GlobalInt64(flags.ArchiveCacheFlag.Name),
	})
}

func jsonGenesisImport(ctx *cli.Context) (err error) {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("this command requires an argument - the genesis file to import")
	}
	if !ctx.IsSet(ExperimentalFlag.Name) {
		return fmt.Errorf("using JSON genesis is for experimental usage only and requires --%s flag", ExperimentalFlag.Name)
	}
	dataDir := ctx.GlobalString(flags.DataDirFlag.Name)
	if dataDir == "" {
		return fmt.Errorf("--%s need to be set", flags.DataDirFlag.Name)
	}
	validatorMode, err := isValidatorModeSet(ctx)
	if err != nil {
		return err
	}
	cacheRatio, err := cacheScaler(ctx)
	if err != nil {
		return err
	}

	genesisJson, err := makefakegenesis.LoadGenesisJson(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("failed to load JSON genesis: %w", err)
	}

	genesisStore, err := makefakegenesis.ApplyGenesisJson(genesisJson)
	if err != nil {
		return fmt.Errorf("failed to prepare JSON genesis: %w", err)
	}
	defer caution.CloseAndReportError(&err, genesisStore, "failed to close the genesis store")
	return genesis.ImportGenesisStore(genesis.ImportParams{
		GenesisStore:     genesisStore,
		DataDir:          dataDir,
		ValidatorMode:    validatorMode,
		CacheRatio:       cacheRatio,
		LiveDbCache:      ctx.GlobalInt64(flags.LiveDbCacheFlag.Name),
		ArchiveCache:     ctx.GlobalInt64(flags.ArchiveCacheFlag.Name),
		StateDbCacheSize: ctx.GlobalInt64(flags.StateDbCacheCapacityFlag.Name),
	})
}

func fakeGenesisImport(ctx *cli.Context) (err error) {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("this command requires an argument - the number of validators in the fake network")
	}
	dataDir := ctx.GlobalString(flags.DataDirFlag.Name)
	if dataDir == "" {
		return fmt.Errorf("failed to read %s, it needs to be set", flags.DataDirFlag.Name)
	}
	validatorsNumber, err := strconv.ParseUint(ctx.Args().First(), 10, 32)
	if err != nil {
		return fmt.Errorf("failed to parse the number of validators: %w", err)
	}
	if validatorsNumber < 1 {
		return fmt.Errorf("the number of validators must be at least 1")
	}
	validatorMode, err := isValidatorModeSet(ctx)
	if err != nil {
		return err
	}
	cacheRatio, err := cacheScaler(ctx)
	if err != nil {
		return err
	}

	var upgrades opera.Upgrades
	upgradesString := ctx.String(FakeUpgrades.Name)
	switch upgradesString {
	case "pano":
		upgrades = opera.GetPanoUpgrades()
	case "allegro":
		upgrades = opera.GetAllegroUpgrades()
	default:
		return fmt.Errorf("invalid profile %v - must be 'pano' or 'allegro'", upgradesString)
	}

	genesisStore := makefakegenesis.FakeGenesisStore(
		idx.Validator(validatorsNumber),
		futils.ToFtm(1_000_000_000),
		futils.ToFtm(5_000_000),
		upgrades,
	)
	defer caution.CloseAndReportError(&err, genesisStore, "failed to close the genesis store")
	return genesis.ImportGenesisStore(genesis.ImportParams{
		GenesisStore:  genesisStore,
		DataDir:       dataDir,
		ValidatorMode: validatorMode,
		CacheRatio:    cacheRatio,
		LiveDbCache:   ctx.GlobalInt64(flags.LiveDbCacheFlag.Name),
		ArchiveCache:  ctx.GlobalInt64(flags.ArchiveCacheFlag.Name),
	})
}

func isValidatorModeSet(ctx *cli.Context) (bool, error) {
	if ctx.IsSet(ModeFlag.Name) {
		mode := ctx.String(ModeFlag.Name)
		if mode != "rpc" && mode != "validator" {
			return false, fmt.Errorf("--%s must be 'rpc' or 'validator'", ModeFlag.Name)
		}
		if mode == "validator" {
			return true, nil
		}
	}
	return false, nil
}

func cacheScaler(ctx *cli.Context) (cachescale.Func, error) {
	targetCache := ctx.GlobalInt(flags.CacheFlag.Name)
	baseSize := db.DefaultCacheSize
	totalMemory := int(memory.TotalMemory() / opt.MiB)
	maxCache := totalMemory * 3 / 5
	if maxCache < baseSize {
		maxCache = baseSize
	}
	if !ctx.GlobalIsSet(flags.CacheFlag.Name) {
		recommendedCache := totalMemory / 2
		if recommendedCache > baseSize {
			log.Warn(fmt.Sprintf("Please add '--%s %d' flag to allocate more cache for the database. Total memory is %d MB.", flags.CacheFlag.Name, recommendedCache, totalMemory))
		}
		return cachescale.Identity, nil
	}
	if targetCache < baseSize {
		return nil, fmt.Errorf("invalid flag %s - minimum cache size is %d MB", flags.CacheFlag.Name, baseSize)
	}
	if totalMemory != 0 && targetCache > maxCache {
		log.Warn(fmt.Sprintf("Requested cache size exceeds 60%% of available memory. Reducing cache size to %d MB.", maxCache))
		targetCache = maxCache
	}

	return cachescale.Ratio{
		Base:   uint64(baseSize - db.ConstantCacheSize),
		Target: uint64(targetCache - db.ConstantCacheSize),
	}, nil
}

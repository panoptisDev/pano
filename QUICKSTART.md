# Pano Blockchain - Quickstart Guide

This guide will help you download, build, and run the Pano blockchain on another computer.

## Prerequisites

- **Go 1.24.0 or higher** ([download](https://go.dev/dl/))
- **Git**
- **protoc** (Protocol Buffers compiler) and **protoc-gen-go** plugin
- **Linux** (tested on Ubuntu, but should work on other distributions)
- At least 4GB RAM available
- 10GB free disk space

## Installation

### 1. Install Go (if not already installed)

```bash
# Download and install Go 1.24.0
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 2. Install protoc and protoc-gen-go

```bash
# Install protoc (Protocol Buffers compiler)
PROTOC_VERSION=28.3
wget https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip
sudo unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local
sudo chmod +x /usr/local/bin/protoc
rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Install protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### 3. Clone and Build Pano

```bash
# Clone the repository
git clone https://github.com/panoptisDev/pano.git
cd pano

# Build the binaries (this will take a few minutes)
make
```

This will create two binaries:
- `build/panod` - the Pano node daemon (67MB)
- `build/panotool` - utility tool for genesis and validator management (66MB)

## Running the Testnet

### Quick Start (3 Validators)

The repository includes everything needed to run a 3-validator testnet:

```bash
# Install Node.js dependencies for transaction testing (optional)
cd testnet
npm install
cd ..

# Start the testnet
cd testnet
./start-testnet.sh
```

This will:
- Start 3 validators on ports 9545-9547 (HTTP RPC) and 30303-30305 (P2P)
- Automatically connect the validators
- Begin producing blocks

### Verify It's Working

```bash
# Check if validators are running
ps aux | grep panod

# Query the blockchain
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://127.0.0.1:9545

# Check current block number
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://127.0.0.1:9545 | jq -r '.result' | xargs printf "%d\n"
```

### Stop the Testnet

```bash
cd testnet
./stop-testnet.sh
```

## Testing Transactions

The testnet includes 5 pre-funded accounts:

- **Validators 1-3**: 1,000,000,000 PANO each
- **User 4**: 100,000 PANO (test account)
- **User 5**: 100,000 PANO (test account)

All private keys are in `testnet/README.md`.

### Send a Test Transaction

```bash
cd testnet

# Send 1000 PANO from User 4 to User 5
node send-transaction.js

# Check balances
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0xE56E6757b8D4124B235436a246af5DCB0a69D14D","latest"],"id":1}' \
  http://127.0.0.1:9545 | jq -r '.result'
```

## Network Configuration

- **Network Name**: pano-testnet-3
- **Network ID**: 4093
- **Gas Token**: PANO
- **Block Time**: ~60 seconds (when empty, faster with transactions)
- **Event Creation**: ~0.6 seconds
- **Max Block Gas**: 5,000,000,000 (5 billion)
- **Min Gas Price**: 1 Gwei

## Understanding the Setup

### Archive Mode

The validators run in **archive mode**, which means they store full blockchain state history. This allows you to:
- Query historical balances
- Get transaction receipts
- Debug contract calls at any block

This is different from validator-only mode which optimizes for consensus but doesn't retain state.

### Consensus

Pano uses **Lachesis consensus**, which provides:
- Fast finality (~0.6 seconds for events)
- Byzantine Fault Tolerance (BFT)
- Asynchronous processing
- No forking

### Gas Fees

Gas fees are **extremely low**:
- ~0.00023 PANO per simple transfer (21,000 gas)
- Gas price: ~10 Gwei (configurable)

## Troubleshooting

### Port Already in Use

If ports 9545-9547 or 30303-30305 are occupied:

1. Edit `testnet/start-testnet.sh`
2. Change the `--http.port` and `--port` values
3. Update `connect-validators.sh` with new RPC ports

### Database Errors

If you see "dirty state" or "non-initialized DB" errors:

```bash
# Clean and reinitialize
cd testnet
./stop-testnet.sh
rm -rf validator*/chaindata validator*/carmen validator*/emitter validator*/p2p
for i in 1 2 3; do
  ../build/panotool --datadir validator$i genesis json --experimental pano-genesis-full.json
done
./start-testnet.sh
```

### Validators Not Connecting

Check the logs:
```bash
tail -f testnet/validator1.log
```

If you see "no suitable peers available", ensure all validators are running and the connect-validators.sh script completed successfully.

## Next Steps

- Read `CHAIN_INFO.md` for detailed chain parameters
- Check `testnet/README.md` for account details
- Explore the genesis file at `testnet/pano-genesis-full.json`
- Deploy smart contracts using the RPC endpoint
- Set up additional validators

## Building on Other Platforms

### macOS

```bash
# Install dependencies
brew install go protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Build (same as Linux)
make
```

### Windows (WSL2)

Use Windows Subsystem for Linux 2 and follow the Linux instructions.

## Support

For issues or questions:
- Check existing issues: https://github.com/panoptisDev/pano/issues
- Create a new issue with logs and steps to reproduce

## License

LGPL-3.0 - See `COPYING.LESSER` for details.

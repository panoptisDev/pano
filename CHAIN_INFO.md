# Pano Blockchain - Chain Information

## Network Details

- **Network Name**: pano-testnet-3
- **Network ID**: 4093
- **Native Token**: PANO
- **Consensus**: Lachesis (DAG-based aBFT)
- **EVM Compatible**: Yes
- **Genesis File**: `testnet/pano-genesis-full.json`

## Performance Metrics

### Consensus Performance

- **Event Creation Time**: ~0.6 seconds (600ms)
  - Events are created continuously by validators
  - Each event can contain 0 or more transactions
  - Events form a Directed Acyclic Graph (DAG)
  - Average processing time: 100-400 microseconds per event

- **Block Finalization Time**: 
  - **With transactions**: Near-instant (sub-second)
  - **Without transactions (empty blocks)**: 60 seconds
  - Controlled by `MaxEmptyBlockSkipPeriod: 60000000000` (60 billion nanoseconds)

### Why Two Different Timings?

The Pano blockchain uses a DAG-based consensus where:
1. **Events** are created every ~0.6 seconds and form the consensus layer
2. **Blocks** are finalized when there are transactions OR after 60 seconds if empty

This design:
- ✅ Provides fast transaction processing (events every 0.6s)
- ✅ Reduces disk space usage (empty blocks only every 60s)
- ✅ Maintains blockchain compatibility (EVM needs blocks)

## Gas Configuration

### Gas Prices (from genesis)

- **Minimum Gas Price**: 1,000,000,000 wei (1 Gwei)
- **Current Gas Price** (observed in tests): ~11,000,000,000 wei (11 Gwei)
- **Base Fee**: 9,999,999,999 wei (initial), adjusts dynamically like EIP-1559

### Gas Limits

- **Max Block Gas**: 5,000,000,000 (5 billion gas)
- **Max Epoch Gas**: 1,875,000,000 (1.875 billion gas)
- **Standard Transfer Gas**: 21,000 gas (same as Ethereum)

### Gas Economics

```
MinGasPrice: 1000000000 wei
GasPowerAllocPerSec: 10000000000
TargetGasPowerPerSecond: 10000000
```

## DAG Configuration

- **Max Parents per Event**: 10
- **Max Free Parents**: 6
- **Max Epoch Duration**: 3,600 seconds (1 hour)

## Transaction Testing Results

### Test Transactions Executed

1. **Transaction 1**: User 4 → User 5
   - Amount: 1,000 PANO
   - Hash: `0xc1007b5c37b7175b427926a738c75fd5f6abcaec357a0d23970f60afb0eeed39`
   - Block: 403
   - Gas Used: 21,000
   - Status: ✅ Confirmed

2. **Transaction 2**: User 5 → User 4
   - Amount: 500 PANO
   - Hash: `0xcdbb07a67ffbafdd7f05086e5478e7093c4b84647b4758888584f71ffa32fd60`
   - Block: 404
   - Gas Used: 21,000
   - Status: ✅ Confirmed

### Transaction Processing

- Transactions are accepted via JSON-RPC (`eth_sendRawTransaction`)
- Included in events within seconds
- Finalized in blocks immediately when present
- Standard Ethereum transaction format (EIP-155 signing)

## Epoch Configuration

- **Epoch Duration**: Up to 3,600 seconds (1 hour)
- **Max Epoch Gas**: 1,875,000,000
- **Current Epoch** (at time of testing): 9

## Validator Configuration

### Active Validators (Testnet)

- **Validator 1**: 0xBcA3d19C24a0ebFc02b4047977F9473C388e4E98
  - Stake: 1,000,000,000 PANO (1 billion)
  - P2P Port: 30303
  - RPC Port: 9545

- **Validator 2**: 0x993669a7793F24b5F2e81c03dB494e0a83EAAE17
  - Stake: 1,000,000,000 PANO (1 billion)
  - P2P Port: 30304
  - RPC Port: 9546

- **Validator 3**: 0x649A72A7c3b30a8a347dC7A549D3e50c3eD4c97c
  - Stake: 1,000,000,000 PANO (1 billion)
  - P2P Port: 30305
  - RPC Port: 9547

## Network Upgrades Enabled

From genesis, the following Ethereum upgrades are enabled:

- ✅ **Homestead** (Block 0)
- ✅ **EIP-150** (Block 0)
- ✅ **EIP-155** (Block 0)
- ✅ **EIP-158** (Block 0)
- ✅ **Byzantium** (Block 0)
- ✅ **Constantinople** (Block 0)
- ✅ **Petersburg** (Block 0)
- ✅ **Istanbul** (Block 0)
- ✅ **Berlin** (Block 0)
- ✅ **London** (Block 0)
- ✅ **Pano** (Custom upgrade, Block 0)

## RPC Endpoints

### HTTP JSON-RPC

- Validator 1: `http://127.0.0.1:9545`
- Validator 2: `http://127.0.0.1:9546`
- Validator 3: `http://127.0.0.1:9547`

### Supported RPC Methods (Tested)

- ✅ `eth_sendRawTransaction` - Send signed transactions
- ✅ `eth_gasPrice` - Get current gas price
- ✅ `eth_feeHistory` - Get fee history
- ⚠️ `eth_getBalance` - Limited (requires archive mode)
- ⚠️ `eth_getTransactionCount` - Limited (requires archive mode)
- ⚠️ `eth_getTransactionByHash` - Limited (requires archive mode)

### Note on Archive Data

The validators are currently running in validator mode, not archive mode. This means:
- ✅ Transactions can be submitted and executed
- ✅ New blocks are produced and finalized
- ❌ Historical state queries (balances, nonces) are limited
- ❌ Transaction receipts may not be retrievable via RPC

To enable full archive functionality, restart validators with archive mode flags.

## Block Structure

### Observed Block Data

```
Block #403:
  - Index: 403
  - Gas Used: 21,000
  - Gas Rate: 737.958
  - Base Fee: 9,999,999,999
  - Transactions: 1
  - Age: 1.210s
  - Epoch: 9

Block #404:
  - Index: 404
  - Gas Used: 21,000
  - Gas Rate: 398.722
  - Base Fee: 8,006,699,725
  - Transactions: 1
  - Age: 1.210s
  - Epoch: 9
```

### Base Fee Adjustment

Like EIP-1559, the base fee adjusts based on block utilization:
- Block 403: 9,999,999,999 wei
- Block 404: 8,006,699,725 wei (~20% decrease)

This demonstrates the dynamic fee mechanism is working.

## Performance Characteristics

### Throughput

- **Theoretical Max**: 5,000,000,000 gas per block / 21,000 gas per tx ≈ **238,095 transactions per block**
- **With 0.6s event time**: ~396,825 tx/second (theoretical maximum)
- **Practical throughput**: Limited by network propagation and validator resources

### Disk Space Efficiency

- Empty blocks created every 60 seconds (vs every 0.6s)
- Saves ~98.3% disk space during low-activity periods
- Blocks with transactions finalize immediately

### Finality

- **Probabilistic finality**: Immediate (DAG consensus)
- **Absolute finality**: After epoch seal
- **Reorganization risk**: Very low (aBFT consensus)

## Comparison to Parent Chain (Sonic)

| Feature | Sonic | Pano (Current Config) |
|---------|-------|----------------------|
| Event Creation | 0.2-0.5s | ~0.6s |
| Block Time (empty) | Sub-second | 60s |
| Block Time (with tx) | ~0.4-1s | Sub-second |
| Consensus | Lachesis aBFT | Lachesis aBFT |
| EVM Compatible | Yes | Yes |
| Network ID | 146 | 4093 |

## Genesis Accounts

See `testnet/README.md` for complete account details including private keys.

### Account Balances (Genesis)

- Validator 1-3: 1,000,000,000 PANO each
- User 4: 100,000 PANO
- User 5: 100,000 PANO

## Testing Summary

✅ **Working Features:**
- Transaction submission and execution
- Block production and finalization
- Gas fee calculation and adjustment
- Validator consensus and P2P networking
- Event creation and DAG consensus
- EVM compatibility

⚠️ **Limited Features:**
- Historical state queries (archive mode required)
- Transaction receipt retrieval
- Balance queries via RPC

## Technical Stack

- **Go Version**: 1.24.0
- **Pano Version**: 2.2.0-dev
- **Fork Base**: Sonic v1.1.3-rc.6
- **Go-Ethereum**: Custom fork (eth1.16.2)
- **Carmen State DB**: v0.0.0-20251004215022-450ae6f398a6
- **Tosca EVM**: v0.0.0-20251004220424-7ee74388a674

## Date of Testing

Tests performed: October 5, 2025

---

*This document reflects the current testnet configuration and observed behavior. Production configuration may differ.*

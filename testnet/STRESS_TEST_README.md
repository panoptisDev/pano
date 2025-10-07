# Pano Blockchain Stress Testing

This directory contains scripts for stress testing the Pano blockchain over extended periods.

## 8-Hour Stress Test

### Overview
The 8-hour stress test sends batches of 10 transactions every 10 minutes for 8 hours, testing the blockchain's stability and performance under consistent load.

### Script: `8hour-stress-test.mjs`

**Configuration:**
- **Duration:** 8 hours
- **Interval:** 10 minutes between batches
- **Transactions per batch:** 10
- **Total transactions:** 480 (10 tx × 48 batches)
- **Transaction amount:** 0.01 PANO each
- **Total PANO sent:** ~4.8 PANO (plus gas fees)

### Prerequisites

1. **Blockchain Running:** Make sure your Pano testnet is running
   ```bash
   cd /home/regium/pano/testnet
   ./start-testnet.sh
   ```

2. **Node.js & Dependencies:** Ensure ethers.js is available
   ```bash
   # If running from smart-contracts directory (has ethers installed)
   cd /home/regium/pano/smart-contracts
   ```

### Running the Test

#### Option 1: From smart-contracts directory (recommended)
```bash
# Navigate to smart-contracts (has ethers.js installed)
cd /home/regium/pano/smart-contracts

# Run the 8-hour stress test
node ../testnet/8hour-stress-test.mjs
```

#### Option 2: Install ethers in testnet directory
```bash
cd /home/regium/pano/testnet
npm install ethers
node 8hour-stress-test.mjs
```

#### Option 3: Run in background with nohup
```bash
cd /home/regium/pano/smart-contracts
nohup node ../testnet/8hour-stress-test.mjs > stress-test-output.log 2>&1 &

# Check the process
ps aux | grep 8hour-stress

# Monitor the log
tail -f stress-test-output.log
```

#### Option 4: Run with tmux/screen (recommended for long tests)
```bash
# Start a tmux session
tmux new -s stress-test

# Run the test
cd /home/regium/pano/smart-contracts
node ../testnet/8hour-stress-test.mjs

# Detach from tmux: Press Ctrl+B, then D
# Reattach later: tmux attach -t stress-test
```

### Output Files

The test creates several output files:

1. **Log File:** `8hour-stress-test-[timestamp].log`
   - Complete transaction log with timestamps
   - Success/failure status for each transaction
   - Balance updates

2. **Stats File:** `8hour-stress-stats.json`
   - Real-time statistics (updated after each batch)
   - Total transactions sent/confirmed/failed
   - Success and confirmation rates
   - Recent batch details

### Monitoring the Test

#### Check Stats File (Live Updates)
```bash
# Watch stats file update in real-time
watch -n 30 'cat /home/regium/pano/testnet/8hour-stress-stats.json'

# Or pretty print with jq
watch -n 30 'cat /home/regium/pano/testnet/8hour-stress-stats.json | jq .'
```

#### Check Log File
```bash
# Follow the log file
tail -f /home/regium/pano/testnet/8hour-stress-test-*.log

# Count successful transactions
grep "✓ TX" /home/regium/pano/testnet/8hour-stress-test-*.log | wc -l

# Count failed transactions
grep "✗ TX" /home/regium/pano/testnet/8hour-stress-test-*.log | wc -l
```

#### Check Sender Balance
```bash
cd /home/regium/pano/smart-contracts
node scripts/interact.js balance 0xE56E6757b8D4124B235436a246af5DCB0a69D14D
```

### Example Stats Output

```json
{
  "test_start": "10/7/2025, 10:30:00 PM",
  "elapsed_hours": 2.5,
  "total_transactions_sent": 150,
  "total_success": 148,
  "total_failed": 2,
  "total_confirmed": 145,
  "success_rate": "98.67%",
  "confirmation_rate": "96.67%",
  "batches_completed": 15,
  "last_update": "10/8/2025, 1:00:00 AM",
  "recent_batches": [
    {
      "batch": 13,
      "timestamp": "10/8/2025, 12:40:00 AM",
      "success": 10,
      "failed": 0,
      "confirmed": 10
    },
    {
      "batch": 14,
      "timestamp": "10/8/2025, 12:50:00 AM",
      "success": 10,
      "failed": 0,
      "confirmed": 9
    },
    {
      "batch": 15,
      "timestamp": "10/8/2025, 1:00:00 AM",
      "success": 10,
      "failed": 0,
      "confirmed": 10
    }
  ]
}
```

### Test Accounts

**Sender (User 4):**
- Address: `0xE56E6757b8D4124B235436a246af5DCB0a69D14D`
- Private Key: `0x5d7a1a73da20b4273a3071411e61e43f46ea8e3cc61f892f72c3bb3b283762da`
- Initial Balance: ~87,160 PANO (check with `node scripts/interact.js balance 0xE56E6757b8D4124B235436a246af5DCB0a69D14D`)

**Recipients (Round-robin):**
- User 5: `0x5Ab49BdE3137bE3e1285319B5F789d9f2831d9B5`
- User 6: `0x8F685ef8A27bF84B6e4Fa27D8D55fD6422596C3D`
- User 6: `0x55FDc3FFF778b5C6be22b1eF791E0d2B451E1fbB`
- User 7: `0x7eF0e79b93c2323bb8Ba24dF90e25ae7a3bda7cB`
- User 8: `0x2Ca7CF4D8EF57F6aC9fb67A1863E2e0851a659bC`
- User 9: `0x83ac78fBDE03e1c2cF8814BF10D92D03cDf3b0cB`
- User 10: `0xA0048E79E8865C72c8de5acf01A90fB0C1BEa5F0`
- User 11: `0x4E97a9Da7e15d4096f9bae8ee475b3fdE8C53D0C`

### Stopping the Test

Press `Ctrl+C` to gracefully stop the test. It will:
1. Stop sending new transactions
2. Save final statistics to JSON file
3. Display a summary

### Expected Results

**Healthy Blockchain:**
- ✅ Success rate: 98%+ 
- ✅ Confirmation rate: 95%+
- ✅ Consistent batch timing (~10 min intervals)
- ✅ No stuck transactions
- ✅ Stable gas prices

**Issues to Watch For:**
- ❌ Success rate below 90%
- ❌ Many unconfirmed transactions
- ❌ Increasing delays between batches
- ❌ Sender running out of PANO

### Customization

Edit the script constants to adjust test parameters:

```javascript
const INTERVAL_MINUTES = 10;    // Change to 5, 15, etc.
const DURATION_HOURS = 8;       // Change to 4, 12, 24, etc.
const TX_PER_BATCH = 10;        // Change batch size
```

### Bash Version (Alternative)

A bash version is also available: `stress-test.sh`

```bash
cd /home/regium/pano/testnet
./stress-test.sh
```

Note: The Node.js version (`8hour-stress-test.mjs`) is recommended as it uses proper transaction signing with ethers.js.

### Troubleshooting

**Problem: "Cannot find module 'ethers'"**
```bash
# Run from smart-contracts directory or install ethers
cd /home/regium/pano/testnet
npm install ethers
```

**Problem: "Error connecting to network"**
```bash
# Make sure blockchain is running
cd /home/regium/pano/testnet
./start-testnet.sh

# Check RPC is accessible
curl -X POST http://127.0.0.1:9545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

**Problem: "Insufficient funds"**
```bash
# Check sender balance
cd /home/regium/pano/smart-contracts
node scripts/interact.js balance 0xE56E6757b8D4124B235436a246af5DCB0a69D14D

# Fund the account if needed (from a validator account)
```

**Problem: Test stops unexpectedly**
```bash
# Check system resources
free -h
df -h

# Check if node is still running
ps aux | grep panod

# Check logs
tail -f /home/regium/pano/testnet/8hour-stress-test-*.log
```

## Other Stress Tests

### Quick Test (10 transactions)
```bash
cd /home/regium/pano/testnet
node stress-test.js
```

### Custom Duration
Modify `DURATION_HOURS` and `INTERVAL_MINUTES` in `8hour-stress-test.mjs` for different test lengths:

- **4-hour test:** `DURATION_HOURS = 4` (24 batches, 240 tx)
- **12-hour test:** `DURATION_HOURS = 12` (72 batches, 720 tx)
- **24-hour test:** `DURATION_HOURS = 24` (144 batches, 1440 tx)

## Analysis After Test

After the test completes, analyze the results:

```bash
# View full stats
cat 8hour-stress-stats.json | jq .

# Count total batches
cat 8hour-stress-test-*.log | grep "Batch #" | wc -l

# Calculate average confirmation time
# (Check the delay between transaction send and confirmation)

# Check blockchain state
cd /home/regium/pano/smart-contracts
node scripts/interact.js balance 0xE56E6757b8D4124B235436a246af5DCB0a69D14D
```

## Performance Metrics to Track

1. **Transaction Success Rate**
2. **Confirmation Rate** 
3. **Average Confirmation Time**
4. **Gas Costs Stability**
5. **Network Uptime**
6. **Memory/CPU Usage** (monitor with `htop`)
7. **Disk I/O** (monitor with `iostat`)

---

**Note:** This stress test is designed for the Pano testnet. Always ensure you have sufficient PANO balance in the sender account before starting the test.

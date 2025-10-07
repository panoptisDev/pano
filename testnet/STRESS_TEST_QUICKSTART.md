# Quick Start: 8-Hour Stress Test

## ✅ Pre-flight Check Complete

**Sender Account Status (User 4):**
- Address: `0xE56E6757b8D4124B235436a246af5DCB0a69D14D`
- Balance: **~87,160 PANO** ✅
- Needed: ~4.8 PANO (480 transactions × 0.01 PANO each)
- **Status: SUFFICIENT FUNDS** ✅

**Environment:**
- Node.js: v22.20.0 ✅
- ethers.js: v6.15.0 ✅
- Blockchain: Running on localhost:9545 ✅

## 🚀 Start the 8-Hour Stress Test

### Recommended Method (using tmux):

```bash
# 1. Start a tmux session
tmux new -s stress8h

# 2. Navigate to smart-contracts directory
cd /home/regium/pano/smart-contracts

# 3. Run the test
node ../testnet/8hour-stress-test.mjs

# 4. Detach from tmux (Ctrl+B, then D)
# The test will continue running in background

# 5. Reattach anytime to check progress
tmux attach -t stress8h
```

### Alternative: Background with nohup

```bash
cd /home/regium/pano/smart-contracts
nohup node ../testnet/8hour-stress-test.mjs > /home/regium/pano/testnet/stress-output.log 2>&1 &

# Get the process ID
echo $!

# Monitor the log
tail -f /home/regium/pano/testnet/stress-output.log
```

## 📊 Monitor Progress

### Check Live Stats (updates every batch):
```bash
# View stats in real-time (every 30 seconds)
watch -n 30 'cat /home/regium/pano/testnet/8hour-stress-stats.json | jq .'
```

### Check Log File:
```bash
# Follow the latest log
tail -f /home/regium/pano/testnet/8hour-stress-test-*.log

# Count successful transactions
grep "✓ TX" /home/regium/pano/testnet/8hour-stress-test-*.log | wc -l
```

### Check Sender Balance:
```bash
cd /home/regium/pano/smart-contracts
node scripts/interact.js balance 0xE56E6757b8D4124B235436a246af5DCB0a69D14D
```

## ⏰ Test Schedule

- **Start:** When you run the script
- **Duration:** 8 hours
- **Batches:** 48 batches total
- **Interval:** 10 minutes between batches
- **Transactions per batch:** 10
- **Total transactions:** 480

### Timeline Example (if started at 10:00 PM):
- 10:00 PM - Batch 1 (10 tx)
- 10:10 PM - Batch 2 (10 tx)
- 10:20 PM - Batch 3 (10 tx)
- ...
- 5:50 AM - Batch 48 (10 tx)
- 6:00 AM - Test complete! ✅

## 🛑 Stop the Test

Press `Ctrl+C` in the terminal (or tmux session) to stop gracefully.

The script will:
1. Stop sending transactions
2. Save final stats to `8hour-stress-stats.json`
3. Display summary

## 📈 Expected Results

**Healthy Blockchain:**
- ✅ Success rate: 98%+
- ✅ Confirmation rate: 95%+
- ✅ Consistent 10-minute intervals
- ✅ No transaction failures

**Sample Output:**
```
[2025-10-07T22:00:00.000Z] [Batch #1] 10/7/2025, 10:00:00 PM
  Starting nonce: 156
  Sender balance: 12285.0 PANO
  ✓ TX #1: 0x1234...abcd -> 0xE56E6757... (nonce: 156)
  ✓ TX #2: 0x5678...efgh -> 0x8F685ef8... (nonce: 157)
  ...
  Batch result: 10 success, 0 failed
  Confirmed: 10/10 transactions

[2025-10-07T22:00:05.000Z] Waiting 10 minutes until next batch...
```

## 📝 Files Created

All output files are saved in `/home/regium/pano/testnet/`:

1. **`8hour-stress-test-[timestamp].log`** - Complete transaction log
2. **`8hour-stress-stats.json`** - Real-time statistics (updated every batch)

## 🔧 Troubleshooting

**If test fails to start:**
```bash
# Make sure blockchain is running
cd /home/regium/pano/testnet
./start-testnet.sh

# Check RPC connectivity
curl -X POST http://127.0.0.1:9545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

**If you see "Cannot find module 'ethers'":**
```bash
# Run from smart-contracts directory (has ethers installed)
cd /home/regium/pano/smart-contracts
node ../testnet/8hour-stress-test.mjs
```

## ✨ Ready to Start!

Everything is configured and ready. Just run:

```bash
tmux new -s stress8h
cd /home/regium/pano/smart-contracts
node ../testnet/8hour-stress-test.mjs
```

Then detach with `Ctrl+B, D` and the test will run for 8 hours automatically! 🚀

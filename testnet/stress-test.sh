#!/bin/bash

# Pano Blockchain Stress Test Script
# Sends 10 transactions every 10 minutes for 8 hours
# Tests chain stability and performance over extended period

# Configuration
RPC_URL="http://127.0.0.1:9545"
CHAIN_ID=4093
INTERVAL_MINUTES=10
DURATION_HOURS=8
TX_PER_BATCH=10

# Test accounts (from testnet/README.md)
SENDER_KEY="0x2a8787659667ab6c826e3d0bf40ab6e0917d0a44e9bc8d4e1db1e5df3c4bad63"
SENDER_ADDR="0x2eF0d698e627724949D202964A4c5e989A186276"

# Recipient addresses (cycle through test accounts)
RECIPIENTS=(
    "0xE56E6757b8D4124B235436a246af5DCB0a69D14D"  # User 4
    "0x8F685ef8A27bF84B6e4Fa27D8D55fD6422596C3D"  # User 5
    "0x55FDc3FFF778b5C6be22b1eF791E0d2B451E1fbB"  # User 6
    "0x7eF0e79b93c2323bb8Ba24dF90e25ae7a3bda7cB"  # User 7
    "0x2Ca7CF4D8EF57F6aC9fb67A1863E2e0851a659bC"  # User 8
    "0x83ac78fBDE03e1c2cF8814BF10D92D03cDf3b0cB"  # User 9
    "0xA0048E79E8865C72c8de5acf01A90fB0C1BEa5F0"  # User 10
    "0x4E97a9Da7e15d4096f9bae8ee475b3fdE8C53D0C"  # User 11
)

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Logging
LOG_FILE="stress-test-$(date +%Y%m%d-%H%M%S).log"
STATS_FILE="stress-test-stats.json"

# Initialize stats
TOTAL_TX_SENT=0
TOTAL_TX_SUCCESS=0
TOTAL_TX_FAILED=0
START_TIME=$(date +%s)

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Pano Blockchain Stress Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Configuration:${NC}"
echo "  RPC URL: $RPC_URL"
echo "  Chain ID: $CHAIN_ID"
echo "  Interval: $INTERVAL_MINUTES minutes"
echo "  Duration: $DURATION_HOURS hours"
echo "  Transactions per batch: $TX_PER_BATCH"
echo "  Sender: $SENDER_ADDR"
echo "  Log file: $LOG_FILE"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop the test${NC}"
echo ""

# Function to get nonce
get_nonce() {
    curl -s -X POST $RPC_URL \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionCount\",\"params\":[\"$SENDER_ADDR\",\"pending\"],\"id\":1}" \
        | grep -o '"result":"[^"]*"' | sed 's/"result":"\(.*\)"/\1/'
}

# Function to get balance
get_balance() {
    local addr=$1
    curl -s -X POST $RPC_URL \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBalance\",\"params\":[\"$addr\",\"latest\"],\"id\":1}" \
        | grep -o '"result":"[^"]*"' | sed 's/"result":"\(.*\)"/\1/'
}

# Function to send transaction
send_transaction() {
    local to_addr=$1
    local nonce_hex=$2
    local amount_wei="0x2386F26FC10000"  # 0.01 PANO
    
    # Build transaction
    local gas_price="0x3B9ACA00"  # 1 Gwei
    local gas_limit="0x5208"      # 21000
    
    # Create raw transaction (simplified - would need proper signing in production)
    local tx_hash=$(curl -s -X POST $RPC_URL \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\":\"2.0\",
            \"method\":\"eth_sendTransaction\",
            \"params\":[{
                \"from\":\"$SENDER_ADDR\",
                \"to\":\"$to_addr\",
                \"gas\":\"$gas_limit\",
                \"gasPrice\":\"$gas_price\",
                \"value\":\"$amount_wei\",
                \"nonce\":\"$nonce_hex\"
            }],
            \"id\":1
        }" | grep -o '"result":"[^"]*"' | sed 's/"result":"\(.*\)"/\1/')
    
    echo "$tx_hash"
}

# Function to check transaction receipt
check_receipt() {
    local tx_hash=$1
    curl -s -X POST $RPC_URL \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionReceipt\",\"params\":[\"$tx_hash\"],\"id\":1}" \
        | grep -q '"status":"0x1"'
    return $?
}

# Function to send batch of transactions
send_batch() {
    local batch_num=$1
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo -e "${BLUE}[Batch #$batch_num] $timestamp${NC}" | tee -a "$LOG_FILE"
    
    # Get current nonce
    local nonce_hex=$(get_nonce)
    local nonce_dec=$((16#${nonce_hex#0x}))
    
    echo "  Starting nonce: $nonce_hex ($nonce_dec)" | tee -a "$LOG_FILE"
    
    local batch_success=0
    local batch_failed=0
    
    # Send TX_PER_BATCH transactions
    for i in $(seq 1 $TX_PER_BATCH); do
        # Select recipient (round-robin)
        local recipient_idx=$(( (batch_num * TX_PER_BATCH + i) % ${#RECIPIENTS[@]} ))
        local recipient=${RECIPIENTS[$recipient_idx]}
        
        # Calculate nonce for this tx
        local current_nonce_dec=$((nonce_dec + i - 1))
        local current_nonce_hex=$(printf "0x%x" $current_nonce_dec)
        
        # Send transaction
        local tx_hash=$(send_transaction "$recipient" "$current_nonce_hex")
        
        if [ -n "$tx_hash" ] && [ "$tx_hash" != "null" ]; then
            echo -e "  ${GREEN}✓${NC} TX #$i: $tx_hash -> ${recipient:0:10}... (nonce: $current_nonce_hex)" | tee -a "$LOG_FILE"
            ((batch_success++))
            ((TOTAL_TX_SUCCESS++))
        else
            echo -e "  ${RED}✗${NC} TX #$i: Failed to send to ${recipient:0:10}... (nonce: $current_nonce_hex)" | tee -a "$LOG_FILE"
            ((batch_failed++))
            ((TOTAL_TX_FAILED++))
        fi
        
        ((TOTAL_TX_SENT++))
        
        # Small delay between transactions
        sleep 0.5
    done
    
    echo "  Batch result: $batch_success success, $batch_failed failed" | tee -a "$LOG_FILE"
    
    # Get updated balance
    local balance_hex=$(get_balance "$SENDER_ADDR")
    local balance_eth=$(echo "scale=4; $((16#${balance_hex#0x})) / 1000000000000000000" | bc)
    echo "  Sender balance: $balance_eth PANO" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
}

# Function to update stats file
update_stats() {
    local current_time=$(date +%s)
    local elapsed=$((current_time - START_TIME))
    local elapsed_hours=$(echo "scale=2; $elapsed / 3600" | bc)
    
    cat > "$STATS_FILE" <<EOF
{
  "test_start": "$(date -d @$START_TIME '+%Y-%m-%d %H:%M:%S')",
  "elapsed_hours": "$elapsed_hours",
  "total_transactions_sent": $TOTAL_TX_SENT,
  "total_success": $TOTAL_TX_SUCCESS,
  "total_failed": $TOTAL_TX_FAILED,
  "success_rate": "$(echo "scale=2; $TOTAL_TX_SUCCESS * 100 / $TOTAL_TX_SENT" | bc 2>/dev/null || echo "0")%",
  "batches_completed": $((TOTAL_TX_SENT / TX_PER_BATCH)),
  "last_update": "$(date '+%Y-%m-%d %H:%M:%S')"
}
EOF
}

# Function to display summary
display_summary() {
    local current_time=$(date +%s)
    local elapsed=$((current_time - START_TIME))
    local elapsed_hours=$(echo "scale=2; $elapsed / 3600" | bc)
    
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   Stress Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo "  Duration: $elapsed_hours hours"
    echo "  Total transactions: $TOTAL_TX_SENT"
    echo -e "  ${GREEN}Success: $TOTAL_TX_SUCCESS${NC}"
    echo -e "  ${RED}Failed: $TOTAL_TX_FAILED${NC}"
    if [ $TOTAL_TX_SENT -gt 0 ]; then
        local success_rate=$(echo "scale=2; $TOTAL_TX_SUCCESS * 100 / $TOTAL_TX_SENT" | bc)
        echo "  Success rate: $success_rate%"
    fi
    echo ""
    echo "  Log file: $LOG_FILE"
    echo "  Stats file: $STATS_FILE"
    echo -e "${BLUE}========================================${NC}"
}

# Cleanup function
cleanup() {
    echo ""
    echo -e "${YELLOW}Stopping stress test...${NC}"
    update_stats
    display_summary
    exit 0
}

# Trap Ctrl+C
trap cleanup SIGINT SIGTERM

# Main loop
TOTAL_BATCHES=$((DURATION_HOURS * 60 / INTERVAL_MINUTES))
echo -e "${GREEN}Starting stress test for $DURATION_HOURS hours ($TOTAL_BATCHES batches)${NC}"
echo ""

for batch in $(seq 1 $TOTAL_BATCHES); do
    send_batch $batch
    update_stats
    
    # Check if this is the last batch
    if [ $batch -lt $TOTAL_BATCHES ]; then
        echo -e "${YELLOW}Waiting $INTERVAL_MINUTES minutes until next batch...${NC}"
        echo ""
        sleep $((INTERVAL_MINUTES * 60))
    fi
done

# Final summary
echo -e "${GREEN}Stress test completed!${NC}"
update_stats
display_summary

exit 0

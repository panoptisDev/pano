#!/bin/bash
# Start Pano 3-Validator Testnet

PANO_ROOT="/home/regium/pano"
cd "$PANO_ROOT"

echo "Starting Pano 3-Validator Testnet..."
echo "========================================"

# Kill any existing panod processes
pkill -9 panod 2>/dev/null

# Clean up old logs
rm -f /tmp/validator*.log

# Start Validator 1
echo "Starting Validator 1 on port 30303, HTTP 9545..."
./build/panod \
  --datadir=testnet/validator1 \
  --port=30303 \
  --http \
  --http.addr=127.0.0.1 \
  --http.port=9545 \
  --http.api=eth,web3,net,pano,admin \
  --mode=validator \
  --fakenet=1/3 \
  --verbosity=3 \
  --maxpeers=10 \
  --nat=none \
  > /tmp/validator1.log 2>&1 &

VAL1_PID=$!
echo "Validator 1 started (PID: $VAL1_PID)"

sleep 2

# Start Validator 2
echo "Starting Validator 2 on port 30304, HTTP 9546..."
./build/panod \
  --datadir=testnet/validator2 \
  --port=30304 \
  --http \
  --http.addr=127.0.0.1 \
  --http.port=9546 \
  --http.api=eth,web3,net,pano,admin \
  --mode=validator \
  --fakenet=2/3 \
  --verbosity=3 \
  --maxpeers=10 \
  --nat=none \
  > /tmp/validator2.log 2>&1 &

VAL2_PID=$!
echo "Validator 2 started (PID: $VAL2_PID)"

sleep 2

# Start Validator 3
echo "Starting Validator 3 on port 30305, HTTP 9547..."
./build/panod \
  --datadir=testnet/validator3 \
  --port=30305 \
  --http \
  --http.addr=127.0.0.1 \
  --http.port=9547 \
  --http.api=eth,web3,net,pano,admin \
  --mode=validator \
  --fakenet=3/3 \
  --verbosity=3 \
  --maxpeers=10 \
  --nat=none \
  > /tmp/validator3.log 2>&1 &

VAL3_PID=$!
echo "Validator 3 started (PID: $VAL3_PID)"

sleep 2

echo ""
echo "========================================"
echo "Testnet Status:"
echo "========================================"
echo "Validator 1: http://127.0.0.1:9545 (PID: $VAL1_PID)"
echo "Validator 2: http://127.0.0.1:9546 (PID: $VAL2_PID)"
echo "Validator 3: http://127.0.0.1:9547 (PID: $VAL3_PID)"
echo ""
echo "Connecting validators..."
./testnet/connect-validators.sh
echo ""
echo "Logs:"
echo "  tail -f /tmp/validator1.log"
echo "  tail -f /tmp/validator2.log"
echo "  tail -f /tmp/validator3.log"
echo ""
echo "To stop: pkill panod"
echo "========================================"

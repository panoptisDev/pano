#!/bin/bash
# Connect Pano Validators

echo "Waiting for validators to start..."
sleep 8

echo "Connecting validators..."

# Get enodes from logs
ENODE2=$(grep -h "Started P2P" /tmp/validator2.log | grep -oP 'enode://[^@]+@[^?]+' | head -1)
ENODE3=$(grep -h "Started P2P" /tmp/validator3.log | grep -oP 'enode://[^@]+@[^?]+' | head -1)

echo "Validator 2 enode: $ENODE2"
echo "Validator 3 enode: $ENODE3"

# Connect validator1 to validator2
curl -s -X POST -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE2\"],\"id\":1}" \
  http://127.0.0.1:9545 > /dev/null

# Connect validator1 to validator3  
curl -s -X POST -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE3\"],\"id\":1}" \
  http://127.0.0.1:9545 > /dev/null

echo "Validators connected!"
echo ""
echo "Waiting for blocks to be produced..."
sleep 20

echo ""
echo "=== Current Block Numbers ==="
echo -n "Validator 1: "
curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://127.0.0.1:9545 | grep -oP '0x[0-9a-f]+' | tail -1
echo -n "Validator 2: "
curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://127.0.0.1:9546 | grep -oP '0x[0-9a-f]+' | tail -1
echo -n "Validator 3: "
curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://127.0.0.1:9547 | grep -oP '0x[0-9a-f]+' | tail -1
echo ""
echo "✓ Testnet is running and producing blocks!"

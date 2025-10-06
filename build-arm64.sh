#!/bin/bash

# Cross-compile Pano for ARM64 (Rock Pi 4 SE)

set -e

echo "=== Cross-Compiling Pano for ARM64 ==="
echo ""

# Set Go path
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go

# Create output directory
OUTPUT_DIR="build/arm64"
mkdir -p "$OUTPUT_DIR"

# Cross-compilation settings
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=1
export CC=aarch64-linux-gnu-gcc
export CXX=aarch64-linux-gnu-g++

echo "Target: $GOOS/$GOARCH"
echo "Output: $OUTPUT_DIR"
echo ""

# Check if cross-compiler is installed
if ! command -v aarch64-linux-gnu-gcc &> /dev/null; then
    echo "❌ ARM64 cross-compiler not found!"
    echo ""
    echo "Install it with:"
    echo "  sudo apt-get update"
    echo "  sudo apt-get install -y gcc-aarch64-linux-gnu g++-aarch64-linux-gnu"
    echo ""
    exit 1
fi

echo "✅ Cross-compiler found: $(aarch64-linux-gnu-gcc --version | head -1)"
echo ""

# Build panod for ARM64
echo "Building panod for ARM64..."
/usr/local/go/bin/go build -v \
    -ldflags "-s -w" \
    -o "$OUTPUT_DIR/panod" \
    ./cmd/panod

echo "✅ panod built successfully"
echo ""

# Build panotool for ARM64
echo "Building panotool for ARM64..."
/usr/local/go/bin/go build -v \
    -ldflags "-s -w" \
    -o "$OUTPUT_DIR/panotool" \
    ./cmd/panotool

echo "✅ panotool built successfully"
echo ""

# Get file info
echo "=== Build Complete ==="
ls -lh "$OUTPUT_DIR/"
echo ""
file "$OUTPUT_DIR/panod"
file "$OUTPUT_DIR/panotool"
echo ""

# Create deployment package
DEPLOY_DIR="deploy-arm64"
rm -rf "$DEPLOY_DIR"
mkdir -p "$DEPLOY_DIR/bin"
mkdir -p "$DEPLOY_DIR/testnet"

# Copy binaries
cp "$OUTPUT_DIR/panod" "$DEPLOY_DIR/bin/"
cp "$OUTPUT_DIR/panotool" "$DEPLOY_DIR/bin/"

# Copy testnet configuration
cp testnet/pano-genesis-full.json "$DEPLOY_DIR/testnet/"
cp testnet/README.md "$DEPLOY_DIR/testnet/"
cp -r testnet/validator1/keystore "$DEPLOY_DIR/testnet/validator1/"
cp -r testnet/validator2/keystore "$DEPLOY_DIR/testnet/validator2/"
cp -r testnet/validator3/keystore "$DEPLOY_DIR/testnet/validator3/"

# Create deployment scripts
cat > "$DEPLOY_DIR/testnet/init-validator.sh" << 'SCRIPT'
#!/bin/bash
# Initialize a validator on Rock Pi

VALIDATOR_NUM=$1
if [ -z "$VALIDATOR_NUM" ]; then
    echo "Usage: ./init-validator.sh <1|2|3>"
    exit 1
fi

DATADIR="validator${VALIDATOR_NUM}"
GENESIS="../testnet/pano-genesis-full.json"

echo "Initializing validator ${VALIDATOR_NUM}..."
../bin/panotool --datadir "$DATADIR" genesis json --experimental "$GENESIS"
echo "✅ Validator ${VALIDATOR_NUM} initialized"
SCRIPT

cat > "$DEPLOY_DIR/testnet/start-validator.sh" << 'SCRIPT'
#!/bin/bash
# Start a validator on Rock Pi

VALIDATOR_NUM=$1
VALIDATOR_IP=$2

if [ -z "$VALIDATOR_NUM" ] || [ -z "$VALIDATOR_IP" ]; then
    echo "Usage: ./start-validator.sh <1|2|3> <IP_ADDRESS>"
    exit 1
fi

# Port configuration
HTTP_PORT=$((9544 + $VALIDATOR_NUM))
P2P_PORT=$((30302 + $VALIDATOR_NUM))

echo "Starting Validator ${VALIDATOR_NUM}"
echo "  HTTP RPC: http://${VALIDATOR_IP}:${HTTP_PORT}"
echo "  P2P: ${VALIDATOR_IP}:${P2P_PORT}"
echo ""

nohup ../bin/panod \
    --datadir=validator${VALIDATOR_NUM} \
    --port=${P2P_PORT} \
    --http \
    --http.addr=0.0.0.0 \
    --http.port=${HTTP_PORT} \
    --http.api=eth,web3,net,pano,admin,debug \
    --fakenet=${VALIDATOR_NUM}/3 \
    --verbosity=3 \
    --maxpeers=10 \
    --nat=extip:${VALIDATOR_IP} \
    --statedb.archivecache=536870912 \
    > validator${VALIDATOR_NUM}.log 2>&1 &

echo $! > validator${VALIDATOR_NUM}.pid
echo "✅ Validator ${VALIDATOR_NUM} started (PID: $(cat validator${VALIDATOR_NUM}.pid))"
echo "   Log: validator${VALIDATOR_NUM}.log"
SCRIPT

cat > "$DEPLOY_DIR/testnet/stop-validator.sh" << 'SCRIPT'
#!/bin/bash
# Stop a validator on Rock Pi

VALIDATOR_NUM=$1
if [ -z "$VALIDATOR_NUM" ]; then
    echo "Usage: ./stop-validator.sh <1|2|3>"
    exit 1
fi

if [ -f "validator${VALIDATOR_NUM}.pid" ]; then
    PID=$(cat "validator${VALIDATOR_NUM}.pid")
    kill $PID 2>/dev/null || true
    rm "validator${VALIDATOR_NUM}.pid"
    echo "✅ Validator ${VALIDATOR_NUM} stopped"
else
    echo "❌ No PID file found for validator ${VALIDATOR_NUM}"
fi
SCRIPT

cat > "$DEPLOY_DIR/testnet/connect-validators.sh" << 'SCRIPT'
#!/bin/bash
# Connect validators on Rock Pi network

VALIDATOR1_IP=$1
VALIDATOR2_IP=$2
VALIDATOR3_IP=$3

if [ -z "$VALIDATOR1_IP" ] || [ -z "$VALIDATOR2_IP" ] || [ -z "$VALIDATOR3_IP" ]; then
    echo "Usage: ./connect-validators.sh <IP1> <IP2> <IP3>"
    exit 1
fi

echo "Connecting validators..."
echo "  Validator 1: $VALIDATOR1_IP:9545"
echo "  Validator 2: $VALIDATOR2_IP:9546"
echo "  Validator 3: $VALIDATOR3_IP:9547"
echo ""

# Get enodes from each validator
ENODE1=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' http://${VALIDATOR1_IP}:9545 | jq -r '.result.enode')
ENODE2=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' http://${VALIDATOR2_IP}:9546 | jq -r '.result.enode')
ENODE3=$(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' http://${VALIDATOR3_IP}:9547 | jq -r '.result.enode')

# Connect Validator 1 to 2 and 3
echo "Connecting Validator 1 to peers..."
curl -s -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE2\"],\"id\":1}" http://${VALIDATOR1_IP}:9545 > /dev/null
curl -s -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE3\"],\"id\":1}" http://${VALIDATOR1_IP}:9545 > /dev/null

# Connect Validator 2 to 3
echo "Connecting Validator 2 to peers..."
curl -s -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE3\"],\"id\":1}" http://${VALIDATOR2_IP}:9546 > /dev/null

echo "✅ Validators connected!"
SCRIPT

chmod +x "$DEPLOY_DIR/testnet/init-validator.sh"
chmod +x "$DEPLOY_DIR/testnet/start-validator.sh"
chmod +x "$DEPLOY_DIR/testnet/stop-validator.sh"
chmod +x "$DEPLOY_DIR/testnet/connect-validators.sh"

# Create README for Rock Pi deployment
cat > "$DEPLOY_DIR/README.md" << 'README'
# Pano Blockchain - Rock Pi 4 SE ARM64 Deployment

This package contains ARM64 binaries and configuration for running Pano validators on Rock Pi 4 SE boards.

## Hardware Requirements (per Rock Pi)

- **Board**: Rock Pi 4 SE
- **RAM**: 4GB (minimum)
- **Storage**: 256GB M.2 SSD
- **Network**: Gigabit Ethernet
- **OS**: Ubuntu 22.04 ARM64 or Debian ARM64

## Rock Pi 4 SE Specifications

- CPU: Rockchip RK3399 (Dual Cortex-A72 @ 1.8GHz + Quad Cortex-A53 @ 1.4GHz)
- Architecture: ARM64 (aarch64)
- Perfect for blockchain validators!

## Quick Start

### 1. Prepare Each Rock Pi

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install required packages
sudo apt-get install -y curl jq

# Create pano directory
mkdir -p ~/pano
cd ~/pano
```

### 2. Transfer Files to Each Rock Pi

From your build machine:

```bash
# Copy to Rock Pi 1
scp -r deploy-arm64/* user@rockpi1:~/pano/

# Copy to Rock Pi 2
scp -r deploy-arm64/* user@rockpi2:~/pano/

# Copy to Rock Pi 3
scp -r deploy-arm64/* user@rockpi3:~/pano/
```

### 3. Initialize Validators

On each Rock Pi:

```bash
cd ~/pano/testnet

# Rock Pi 1 (Validator 1)
./init-validator.sh 1

# Rock Pi 2 (Validator 2)
./init-validator.sh 2

# Rock Pi 3 (Validator 3)
./init-validator.sh 3
```

### 4. Start Validators

On each Rock Pi (replace IP addresses with your actual IPs):

```bash
# Rock Pi 1
./start-validator.sh 1 192.168.1.101

# Rock Pi 2
./start-validator.sh 2 192.168.1.102

# Rock Pi 3
./start-validator.sh 3 192.168.1.103
```

### 5. Connect Validators (run on any Rock Pi)

```bash
./connect-validators.sh 192.168.1.101 192.168.1.102 192.168.1.103
```

## Storage Configuration

### Mount M.2 SSD (on each Rock Pi)

```bash
# Find your M.2 drive
lsblk

# Format (if new)
sudo mkfs.ext4 /dev/nvme0n1

# Create mount point
sudo mkdir -p /mnt/pano-data

# Mount
sudo mount /dev/nvme0n1 /mnt/pano-data

# Auto-mount on boot
echo '/dev/nvme0n1 /mnt/pano-data ext4 defaults 0 0' | sudo tee -a /etc/fstab

# Move validator data to SSD
sudo mkdir -p /mnt/pano-data/validator1
sudo chown $USER:$USER /mnt/pano-data/validator1
ln -s /mnt/pano-data/validator1 ~/pano/testnet/validator1
```

## Network Configuration

### Port Forwarding (if behind router)

For each Rock Pi, forward these ports:

**Rock Pi 1:**
- TCP 9545 (HTTP RPC)
- TCP/UDP 30303 (P2P)

**Rock Pi 2:**
- TCP 9546 (HTTP RPC)
- TCP/UDP 30304 (P2P)

**Rock Pi 3:**
- TCP 9547 (HTTP RPC)
- TCP/UDP 30305 (P2P)

### Static IPs

Recommended: Set static IPs for each Rock Pi in your router.

## Monitoring

### Check Validator Status

```bash
# Check if running
ps aux | grep panod

# Check logs
tail -f validator1.log

# Check current block
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:9545 | jq -r '.result' | xargs printf "%d\n"

# Check peers
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' \
  http://localhost:9545 | jq
```

### System Resources

```bash
# CPU and Memory
htop

# Disk usage
df -h /mnt/pano-data

# Network
iftop
```

## Performance Optimization

### For Rock Pi 4 SE

```bash
# Enable performance governor
echo performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

# Increase file descriptors
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Optimize network
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.wmem_max=134217728
```

### Archive Cache Optimization

The validators are configured with 512MB archive cache (suitable for 4GB RAM).

To adjust:
- Edit `start-validator.sh`
- Change `--statedb.archivecache=536870912` (bytes)
- Default: 512MB (536870912 bytes)
- Maximum for 4GB RAM: 1GB (1073741824 bytes)

## Troubleshooting

### Low Memory

If you experience OOM errors:

1. Reduce archive cache in `start-validator.sh`
2. Add swap:
```bash
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

### Slow Performance

1. Check CPU governor is set to "performance"
2. Ensure M.2 SSD is used for data storage
3. Monitor disk I/O: `iostat -x 1`

### Network Issues

1. Check firewall: `sudo ufw status`
2. Allow ports: `sudo ufw allow 9545/tcp && sudo ufw allow 30303`
3. Verify connectivity: `nc -zv <other-rockpi-ip> 30303`

## Backup

### Backup Keystores (Important!)

```bash
# On each Rock Pi
tar -czf keystore-backup.tar.gz testnet/validator*/keystore
scp keystore-backup.tar.gz user@backup-server:~/backups/
```

### Backup Database (Optional)

```bash
# Stop validator first
./stop-validator.sh 1

# Backup
tar -czf validator1-backup.tar.gz validator1/

# Restart
./start-validator.sh 1 <IP>
```

## MetaMask Access

You can access the testnet from your PC using MetaMask:

**RPC URL**: `http://<rockpi-ip>:9545`
**Chain ID**: 4093
**Currency**: PANO

See `testnet/METAMASK_SETUP.md` for account details.

## Long-term Operation

### Auto-restart on Boot

Create systemd service on each Rock Pi:

```bash
sudo tee /etc/systemd/system/pano-validator.service << 'EOF'
[Unit]
Description=Pano Validator
After=network.target

[Service]
Type=forking
User=youruser
WorkingDirectory=/home/youruser/pano/testnet
ExecStart=/home/youruser/pano/testnet/start-validator.sh 1 192.168.1.101
ExecStop=/home/youruser/pano/testnet/stop-validator.sh 1
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable pano-validator
sudo systemctl start pano-validator
```

### Log Rotation

```bash
sudo tee /etc/logrotate.d/pano << 'EOF'
/home/*/pano/testnet/validator*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
EOF
```

## Estimated Storage Usage

With 256GB M.2 SSD per validator:

- **Initial**: ~1GB (genesis + databases)
- **Per day**: ~500MB (blocks + state)
- **Capacity**: ~500 days of operation
- **Recommendation**: Monitor and prune after 6 months

## Support

For issues specific to Rock Pi 4 SE deployment, check:
- Rock Pi Wiki: https://wiki.radxa.com/Rockpi4
- Pano Issues: https://github.com/panoptisDev/pano/issues

## Summary

✅ ARM64 optimized binaries
✅ Distributed validator setup
✅ M.2 SSD support
✅ Auto-start configuration
✅ Network discovery
✅ Remote MetaMask access
✅ Long-term operation support

**Your Rock Pi 4 SE cluster is ready to run a production Pano testnet!**
README

echo "✅ Deployment package created: $DEPLOY_DIR/"
echo ""
echo "Package contents:"
du -sh "$DEPLOY_DIR"
ls -la "$DEPLOY_DIR/"
echo ""
echo "Next steps:"
echo "  1. Install ARM64 cross-compiler (if not done):"
echo "     sudo apt-get install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu"
echo ""
echo "  2. This script will create ARM64 binaries in: $OUTPUT_DIR/"
echo ""
echo "  3. Copy $DEPLOY_DIR/ to your Rock Pi boards"
echo ""
echo "  4. Follow instructions in $DEPLOY_DIR/README.md"

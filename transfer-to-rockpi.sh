#!/bin/bash

# Transfer deployment package to Rock Pi 4 SE boards
# Usage: ./transfer-to-rockpi.sh <username> <rockpi1-ip> <rockpi2-ip> <rockpi3-ip>

set -e

# Check arguments
if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <username> <rockpi1-ip> <rockpi2-ip> <rockpi3-ip>"
    echo ""
    echo "Example:"
    echo "  $0 rockpi 192.168.1.101 192.168.1.102 192.168.1.103"
    exit 1
fi

USERNAME=$1
ROCKPI1_IP=$2
ROCKPI2_IP=$3
ROCKPI3_IP=$4

DEPLOY_DIR="deploy-arm64"

if [ ! -d "$DEPLOY_DIR" ]; then
    echo "❌ Error: $DEPLOY_DIR directory not found!"
    echo "Please run ./build-arm64.sh first"
    exit 1
fi

echo "=== Transferring Pano Deployment to Rock Pi Boards ==="
echo ""
echo "Source: $DEPLOY_DIR"
echo "Target User: $USERNAME"
echo "Rock Pi 1: $ROCKPI1_IP"
echo "Rock Pi 2: $ROCKPI2_IP"
echo "Rock Pi 3: $ROCKPI3_IP"
echo ""

# Function to transfer to one Rock Pi
transfer_to_rockpi() {
    local BOARD_NUM=$1
    local IP=$2
    
    echo "📦 Transferring to Rock Pi $BOARD_NUM ($IP)..."
    
    # Create directory on remote
    ssh "$USERNAME@$IP" "mkdir -p ~/pano" || {
        echo "❌ Failed to connect to Rock Pi $BOARD_NUM at $IP"
        return 1
    }
    
    # Transfer files
    scp -r "$DEPLOY_DIR"/* "$USERNAME@$IP:~/pano/" || {
        echo "❌ Failed to transfer files to Rock Pi $BOARD_NUM"
        return 1
    }
    
    # Set permissions
    ssh "$USERNAME@$IP" "chmod +x ~/pano/bin/* ~/pano/testnet/*.sh" || {
        echo "⚠️  Warning: Could not set execute permissions on Rock Pi $BOARD_NUM"
    }
    
    echo "✅ Transfer to Rock Pi $BOARD_NUM complete"
    echo ""
}

# Transfer to all Rock Pis
echo "Starting transfers..."
echo ""

transfer_to_rockpi 1 "$ROCKPI1_IP"
transfer_to_rockpi 2 "$ROCKPI2_IP"
transfer_to_rockpi 3 "$ROCKPI3_IP"

echo "=== Transfer Complete ==="
echo ""
echo "Next steps on each Rock Pi:"
echo ""
echo "1. Initialize validator:"
echo "   cd ~/pano/testnet"
echo "   ./init-validator.sh <1|2|3>"
echo ""
echo "2. Start validator with IP:"
echo "   ./start-validator.sh <1|2|3> <ip-address>"
echo ""
echo "3. Connect validators (run on any one Rock Pi):"
echo "   ./connect-validators.sh $ROCKPI1_IP $ROCKPI2_IP $ROCKPI3_IP"
echo ""
echo "See ROCK_PI_DEPLOYMENT.md for detailed instructions!"

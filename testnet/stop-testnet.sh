#!/bin/bash
# Stop Pano Testnet

echo "Stopping Pano testnet validators..."
pkill -9 panod

echo "Cleaning up..."
rm -f /tmp/validator*.log

echo "All validators stopped."

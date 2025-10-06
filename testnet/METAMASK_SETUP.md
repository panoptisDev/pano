# MetaMask Setup for Pano Testnet

## Network Configuration

You can connect MetaMask to your local Pano testnet and send transactions manually!

### Add Network to MetaMask

1. Open MetaMask
2. Click on the network dropdown (top of MetaMask)
3. Click "Add Network" → "Add a network manually"
4. Enter the following details:

```
Network Name: Pano Testnet
RPC URL: http://127.0.0.1:9545
Chain ID: 4093
Currency Symbol: PANO
Block Explorer URL: (leave empty)
```

5. Click "Save"

### Import Test Accounts

You can import the test accounts using their private keys:

**⚠️ WARNING: These are TEST ACCOUNTS ONLY - Never use these private keys on mainnet!**

#### User 4 (Currently has ~95,350 PANO)
```
Private Key: 0x5d7a1a73da20b4273a3071411e61e43f46ea8e3cc61f892f72c3bb3b283762da
Address: 0xE56E6757b8D4124B235436a246af5DCB0a69D14D
```

#### User 5 (Currently has ~101,600 PANO)
```
Private Key: 0x5152c0b669f29ae1911ef16a597097709d963b99b13ab5c3632881c893c8be4e
Address: 0x5Ab49BdE3137bE3e1285319B5F789d9f2831d9B5
```

#### To Import:
1. In MetaMask, click the account icon
2. Select "Import Account"
3. Paste the private key
4. Click "Import"

### Validator Accounts (1 Billion PANO each)

These accounts require unlocking the keystore files (password: "password"):

#### Validator 1
```
Address: 0xBcA3d19C24a0ebFc02b4047977F9473C388e4E98
Balance: 1,000,000,000 PANO
Keystore: testnet/validator1/keystore/validator/c0048d505...
```

#### Validator 2
```
Address: 0x993669a7793F24b5F2e81c03dB494e0a83EAAE17
Balance: 1,000,000,000 PANO
Keystore: testnet/validator2/keystore/validator/c0043b406...
```

#### Validator 3
```
Address: 0x649A72A7c3b30a8a347dC7A549D3e50c3eD4c97c
Balance: 1,000,000,000 PANO
Keystore: testnet/validator3/keystore/validator/c0045a463...
```

## Fee Structure (Verified)

Based on the latest test:

- **Gas Limit**: 21,000 (standard transfer)
- **Gas Price**: ~3.17 Gwei (dynamic, based on network)
- **Transaction Fee**: ~0.0000666 PANO
- **Fee as % of 100 PANO transfer**: 0.000067%

### Example Transaction:
```
Sent: 100 PANO
Gas Fee: 0.000066564972159 PANO
Total Cost: 100.000066564972159 PANO
```

## RPC Endpoints

Your local testnet exposes multiple RPC endpoints:

- **Validator 1**: http://127.0.0.1:9545
- **Validator 2**: http://127.0.0.1:9546
- **Validator 3**: http://127.0.0.1:9547

All support the same APIs:
- `eth_*` - Standard Ethereum JSON-RPC
- `web3_*` - Web3 methods
- `net_*` - Network methods
- `pano_*` - Pano-specific methods
- `admin_*` - Admin methods
- `debug_*` - Debug methods

## Using MetaMask

### Send a Transaction:

1. Make sure you're connected to "Pano Testnet"
2. Select the account you want to send from (User 4 or User 5)
3. Click "Send"
4. Enter recipient address (e.g., User 5's address)
5. Enter amount (e.g., 50 PANO)
6. Review gas fee (should be ~0.000066 PANO)
7. Click "Confirm"

### Check Transaction:

After sending, you can verify in the terminal:

```bash
# Get latest block
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://127.0.0.1:9545 | jq

# Get transaction receipt (replace TX_HASH)
curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["TX_HASH"],"id":1}' \
  http://127.0.0.1:9545 | jq
```

## Remote Access (Optional)

If you want to access the RPC from another computer on your network:

1. Edit `testnet/start-testnet.sh`
2. Change `--http.addr=127.0.0.1` to `--http.addr=0.0.0.0`
3. Restart the testnet
4. Use `http://YOUR_IP:9545` in MetaMask

**⚠️ Security Warning**: Only do this on a trusted local network. Never expose RPC to the internet without proper security!

## Troubleshooting

### MetaMask shows "Internal JSON-RPC error"
- Make sure the testnet is running: `ps aux | grep panod`
- Check the RPC endpoint is accessible: `curl http://127.0.0.1:9545`

### Transaction stuck as "Pending"
- Check if validators are producing blocks
- Wait ~60 seconds (block time when empty)
- Check validator logs: `tail -f testnet/validator1.log`

### Wrong nonce error
- Reset account in MetaMask: Settings → Advanced → Clear activity tab data

### Gas price too low
- The network has a minimum gas price of 1 Gwei
- MetaMask should auto-detect the correct gas price
- Manual override: Set gas price to at least 3 Gwei

## Current Network Stats

Run this to see live stats:

```bash
cd /home/regium/pano/testnet
node -e "
const { ethers } = require('ethers');
const provider = new ethers.JsonRpcProvider('http://127.0.0.1:9545');
(async () => {
  const block = await provider.getBlockNumber();
  const gasPrice = await provider.getFeeData();
  console.log('Current Block:', block);
  console.log('Gas Price:', ethers.formatUnits(gasPrice.gasPrice, 'gwei'), 'Gwei');
})();
"
```

## Test Script Available

Use the fee structure test script:

```bash
cd /home/regium/pano/testnet
node test-fee-structure.js
```

This will send a test transaction and show detailed fee breakdown.

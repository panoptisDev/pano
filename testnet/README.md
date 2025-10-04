# Pano 3-Validator Testnet

## Generated Accounts

### Validators (1,000,000,000 PANO each)
1. **Validator 1**
   - Address: `0xBcA3d19C24a0ebFc02b4047977F9473C388e4E98`
   - Private Key: `0xd3d9f821e5df04816f0d97253c1d0a071de76c743d761e703fc50f638d6f5a27`
   - Keystore: `validator1/keystore/`

2. **Validator 2**
   - Address: `0x993669a7793F24b5F2e81c03dB494e0a83EAAE17`
   - Private Key: `0xb766e346d576615d4c476773be40f8fa98ebf458b0165f93aa57566bbab4ca03`
   - Keystore: `validator2/keystore/`

3. **Validator 3**
   - Address: `0x649A72A7c3b30a8a347dC7A549D3e50c3eD4c97c`
   - Private Key: `0x62f2b6e05996548b52966738e3485bed674e339e56a6e56cc55e2572f0371674`
   - Keystore: `validator3/keystore/`

### User Accounts (100,000 PANO each)
4. **User 4**
   - Address: `0xE56E6757b8D4124B235436a246af5DCB0a69D14D`
   - Private Key: `0x5d7a1a73da20b4273a3071411e61e43f46ea8e3cc61f892f72c3bb3b283762da`

5. **User 5**
   - Address: `0x5Ab49BdE3137bE3e1285319B5F789d9f2831d9B5`
   - Private Key: `0x5152c0b669f29ae1911ef16a597097709d963b99b13ab5c3632881c893c8be4e`

## Files

- `pano-genesis-full.json` - Complete genesis file with contract bytecode
- `validator1/`, `validator2/`, `validator3/` - Validator datadirs with keystores
- `validators/keystore/` - All keystores in one directory
- `make_keystores.go` - Tool to regenerate keystores
- `generate_accounts.go` - Tool to generate new accounts

## Network Configuration

- **Network Name**: pano-testnet-3
- **Network ID**: 4093
- **Gas Token**: PANO

## Next Steps

1. **Build panod** (if not already built):
   ```bash
   cd /home/regium/pano
   /usr/local/go/bin/go build -o build/panod ./cmd/panod
   ```

2. **Initialize Validator Datadirs** (if using panotool):
   ```bash
   ./build/panotool genesis import testnet/pano-genesis-full.json --datadir testnet/validator1
   ./build/panotool genesis import testnet/pano-genesis-full.json --datadir testnet/validator2
   ./build/panotool genesis import testnet/pano-genesis-full.json --datadir testnet/validator3
   ```

3. **Start Validators** (example - adjust flags as needed):
   ```bash
   # Validator 1
   ./build/panod --datadir testnet/validator1 --networkid 4093 --port 30303 --http.port 8545 --unlock 0xBcA3d19C24a0ebFc02b4047977F9473C388e4E98 --password <(echo "") &

   # Validator 2  
   ./build/panod --datadir testnet/validator2 --networkid 4093 --port 30304 --http.port 8546 --unlock 0x993669a7793F24b5F2e81c03dB494e0a83EAAE17 --password <(echo "") &

   # Validator 3
   ./build/panod --datadir testnet/validator3 --networkid 4093 --port 30305 --http.port 8547 --unlock 0x649A72A7c3b30a8a347dC7A549D3e50c3eD4c97c --password <(echo "") &
   ```

## Security Warning

⚠️ **DO NOT USE THESE ACCOUNTS ON MAINNET!** 
These are test accounts with publicly known private keys. Only use for local testing.

All keystores are encrypted with an empty password for testing convenience.

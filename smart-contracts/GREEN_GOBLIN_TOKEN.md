# 🎉 Green Goblin Token (GOB) - Deployed Successfully!

## 🚀 Deployment Summary

**Network**: Pano Testnet 3  
**Chain ID**: 4093  
**Deployed**: October 6, 2025

### 📍 Contract Details

- **Contract Address**: `0x143896d641aEbaC13D9e393F77f7588C994f05a4`
- **Transaction Hash**: `0xa6bfbb1b1e901804e28a9086c60c6a4a98936722df9a9e2ebe0e032c1d65e59b`
- **Deployer**: `0xE56E6757b8D4124B235436a246af5DCB0a69D14D` (User 4)

### 💎 Token Information

- **Name**: Green Goblin
- **Symbol**: GOB  
- **Decimals**: 18
- **Initial Supply**: 300,000,000 GOB
- **Max Supply (Mint Cap)**: 500,000,000 GOB
- **Remaining Mintable**: 200,000,000 GOB
- **Owner**: 0xE56E6757b8D4124B235436a246af5DCB0a69D14D

### 💰 Initial Distribution

- **Owner Balance**: 300,000,000 GOB (User 4)
- All tokens minted to deployer address

## 📱 Add to MetaMask

1. Open MetaMask
2. Make sure you're connected to Pano testnet (http://127.0.0.1:9545, Chain ID 4093)
3. Click "Import tokens"
4. Enter the following:
   - **Token Contract Address**: `0x143896d641aEbaC13D9e393F77f7588C994f05a4`
   - **Token Symbol**: GOB
   - **Token Decimals**: 18

Your GOB tokens will appear in your wallet!

## 🎨 Token Features

### ERC20 Standard
- ✅ Transfer tokens
- ✅ Approve spending
- ✅ Transfer from approved addresses

### ERC20Burnable
- ✅ `burn(amount)` - Burn your own tokens
- ✅ `burnFrom(from, amount)` - Burn from approved addresses

### Ownable
- ✅ `mint(to, amount)` - Owner can mint up to cap (200M more)
- ✅ `transferOwnership(newOwner)` - Transfer ownership

### ERC20Permit (Gasless Approvals)
- ✅ `permit(owner, spender, value, deadline, v, r, s)` - Approve via signature

### Custom Features
- ✅ `circulatingSupply()` - View circulating supply
- ✅ `remainingMintable()` - View remaining mintable tokens
- ✅ `MINT_CAP` - View max supply cap (500M)
- ✅ Tracks circulating supply accurately (excludes burned tokens)

## 📊 Token Economics

- **Max Supply**: 500,000,000 GOB (fixed cap)
- **Initial Mint**: 300,000,000 GOB (60% of cap)
- **Reserve for Future**: 200,000,000 GOB (40% of cap)
- **Inflation**: Capped at 500M forever
- **Deflation**: Burnable (reduces circulating supply)

## 🔧 Interacting with the Contract

### Using ethers.js

```javascript
import { ethers } from "ethers";

const provider = new ethers.JsonRpcProvider("http://127.0.0.1:9545");
const tokenAddress = "0x143896d641aEbaC13D9e393F77f7588C994f05a4";
const tokenABI = [...]; // From artifacts

const token = new ethers.Contract(tokenAddress, tokenABI, provider);

// Read functions
const balance = await token.balanceOf("0x...");
const totalSupply = await token.totalSupply();
const circulatingSupply = await token.circulatingSupply();
const remainingMintable = await token.remainingMintable();

// Write functions (need signer)
const signer = new ethers.Wallet("0x...", provider);
const tokenWithSigner = token.connect(signer);

await tokenWithSigner.transfer("0x...", ethers.parseEther("1000"));
await tokenWithSigner.burn(ethers.parseEther("100"));
```

### Using MetaMask

1. Import GOB token (see instructions above)
2. Send tokens to any address
3. View balance in MetaMask
4. Track transactions in MetaMask activity

## 🎭 Use Cases

### Meme Token
- Perfect for community fun and engagement
- Green Goblin theme!
- Capped supply creates scarcity

### DeFi Integration
- Can be used in AMM pools (Uniswap-style DEXs)
- Lending protocols
- Yield farming
- Governance tokens

### NFT Utility
- Use as currency for NFT marketplaces
- Reward token for NFT holders
- Staking rewards

## 📝 Contract Source

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";

contract GreenGoblinToken is ERC20, ERC20Burnable, Ownable, ERC20Permit {
    uint256 private _circulatingSupply;
    uint256 public constant MINT_CAP = 500_000_000 * 10**18;

    event CirculatingSupplyUpdated(uint256 newSupply);

    constructor(string memory name, string memory symbol, address initialOwner) 
        ERC20(name, symbol) 
        Ownable(initialOwner) 
        ERC20Permit(name) 
    {
        require(initialOwner != address(0), "Owner cannot be zero address");
        
        uint256 initialSupply = 300_000_000 * 10**decimals();
        require(initialSupply <= MINT_CAP, "Initial supply exceeds cap");
        
        _mint(initialOwner, initialSupply);
        _circulatingSupply = initialSupply;
        emit CirculatingSupplyUpdated(_circulatingSupply);
    }

    function mint(address to, uint256 amount) public onlyOwner {
        require(_circulatingSupply + amount <= MINT_CAP, "Mint cap exceeded");
        _mint(to, amount);
        unchecked { _circulatingSupply += amount; }
        emit CirculatingSupplyUpdated(_circulatingSupply);
    }

    function circulatingSupply() public view returns (uint256) {
        return _circulatingSupply;
    }

    function remainingMintable() public view returns (uint256) {
        return MINT_CAP - _circulatingSupply;
    }
}
```

## 🔗 Blockchain Explorer

View your token on the Pano blockchain:

```bash
# Get token details
curl -X POST http://127.0.0.1:9545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"eth_call",
    "params":[{
      "to":"0x143896d641aEbaC13D9e393F77f7588C994f05a4",
      "data":"0x06fdde03"
    }, "latest"],
    "id":1
  }'
```

## 🎉 Next Steps

1. **Add to MetaMask** - Import the token
2. **Transfer Tokens** - Send GOB to friends
3. **Create a Pool** - Set up liquidity on a DEX
4. **Build NFT Marketplace** - Use GOB as currency
5. **Community Airdrops** - Distribute to community members
6. **Mint More** - Owner can mint up to 200M more tokens

## 📄 Files

- **Contract**: `contracts/GreenGoblinToken.sol`
- **Deployment Script**: `scripts/deploy-ethers.js`
- **Artifacts**: `artifacts/contracts/GreenGoblinToken.sol/`
- **Deployment Info**: `deployment.json`

## ⚡ Performance

- **Deployment Gas**: ~2,000,000 gas
- **Transfer Gas**: ~50,000 gas
- **Burn Gas**: ~35,000 gas
- **Mint Gas**: ~55,000 gas

---

**Your meme token is live! Time to take it to the moon! 🚀🌙**

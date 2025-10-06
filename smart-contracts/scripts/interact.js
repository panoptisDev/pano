import { ethers } from "ethers";
import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { dirname, join } from "path";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const TOKEN_ADDRESS = "0x143896d641aEbaC13D9e393F77f7588C994f05a4";
const RPC_URL = "http://127.0.0.1:9545";

async function main() {
  const args = process.argv.slice(2);
  const command = args[0];

  if (!command) {
    console.log(`
🎮 Green Goblin Token (GOB) Interaction Script

Usage: node interact.js <command> [options]

Commands:
  balance <address>              - Get GOB balance of an address
  transfer <to> <amount>         - Transfer GOB tokens (requires private key in env)
  mint <to> <amount>             - Mint new GOB tokens (owner only)
  burn <amount>                  - Burn your GOB tokens
  info                           - Get token info
  supply                         - Get supply information
  help                           - Show this help

Environment Variables:
  PRIVATE_KEY - Your private key for signing transactions

Examples:
  node interact.js info
  node interact.js balance 0xE56E6757b8D4124B235436a246af5DCB0a69D14D
  PRIVATE_KEY=0x5d7a... node interact.js transfer 0x7E5F... 1000
  PRIVATE_KEY=0x5d7a... node interact.js mint 0x7E5F... 50000000
  PRIVATE_KEY=0x5d7a... node interact.js burn 1000
    `);
    return;
  }

  // Connect to provider
  const provider = new ethers.JsonRpcProvider(RPC_URL);

  // Load contract ABI
  const artifactPath = join(__dirname, "../artifacts/contracts/GreenGoblinToken.sol/GreenGoblinToken.json");
  const artifact = JSON.parse(readFileSync(artifactPath, "utf8"));

  const token = new ethers.Contract(TOKEN_ADDRESS, artifact.abi, provider);

  switch (command) {
    case "info":
      await getTokenInfo(token);
      break;

    case "supply":
      await getSupplyInfo(token);
      break;

    case "balance":
      if (!args[1]) {
        console.error("❌ Error: Please provide an address");
        return;
      }
      await getBalance(token, args[1]);
      break;

    case "transfer":
      if (!args[1] || !args[2]) {
        console.error("❌ Error: Please provide recipient address and amount");
        return;
      }
      await transfer(token, provider, args[1], args[2]);
      break;

    case "mint":
      if (!args[1] || !args[2]) {
        console.error("❌ Error: Please provide recipient address and amount");
        return;
      }
      await mint(token, provider, args[1], args[2]);
      break;

    case "burn":
      if (!args[1]) {
        console.error("❌ Error: Please provide amount to burn");
        return;
      }
      await burn(token, provider, args[1]);
      break;

    case "help":
      process.argv = process.argv.slice(0, 2); // Reset args
      main();
      break;

    default:
      console.error(`❌ Unknown command: ${command}`);
      console.log("Run 'node interact.js help' for usage");
  }
}

async function getTokenInfo(token) {
  console.log("\n📊 Green Goblin Token (GOB) Information\n");
  
  const name = await token.name();
  const symbol = await token.symbol();
  const decimals = await token.decimals();
  const totalSupply = await token.totalSupply();
  const owner = await token.owner();
  const mintCap = await token.MINT_CAP();

  console.log(`  Name: ${name}`);
  console.log(`  Symbol: ${symbol}`);
  console.log(`  Decimals: ${decimals}`);
  console.log(`  Contract Address: ${TOKEN_ADDRESS}`);
  console.log(`  Owner: ${owner}`);
  console.log(`  Total Supply: ${ethers.formatEther(totalSupply)} GOB`);
  console.log(`  Mint Cap: ${ethers.formatEther(mintCap)} GOB`);
  console.log("");
}

async function getSupplyInfo(token) {
  console.log("\n💰 Supply Information\n");
  
  const totalSupply = await token.totalSupply();
  const circulatingSupply = await token.circulatingSupply();
  const mintCap = await token.MINT_CAP();
  const remainingMintable = await token.remainingMintable();

  console.log(`  Total Supply: ${ethers.formatEther(totalSupply)} GOB`);
  console.log(`  Circulating Supply: ${ethers.formatEther(circulatingSupply)} GOB`);
  console.log(`  Mint Cap: ${ethers.formatEther(mintCap)} GOB`);
  console.log(`  Remaining Mintable: ${ethers.formatEther(remainingMintable)} GOB`);
  console.log(`  % of Cap Minted: ${(Number(circulatingSupply) / Number(mintCap) * 100).toFixed(2)}%`);
  console.log("");
}

async function getBalance(token, address) {
  console.log(`\n💼 Checking balance for ${address}\n`);
  
  try {
    const balance = await token.balanceOf(address);
    console.log(`  Balance: ${ethers.formatEther(balance)} GOB`);
    console.log("");
  } catch (error) {
    console.error(`❌ Error: ${error.message}`);
  }
}

async function transfer(token, provider, to, amount) {
  const privateKey = process.env.PRIVATE_KEY;
  if (!privateKey) {
    console.error("❌ Error: PRIVATE_KEY environment variable not set");
    return;
  }

  const wallet = new ethers.Wallet(privateKey, provider);
  const tokenWithSigner = token.connect(wallet);

  console.log(`\n📤 Transferring ${amount} GOB to ${to}\n`);

  try {
    const amountWei = ethers.parseEther(amount);
    const tx = await tokenWithSigner.transfer(to, amountWei);
    
    console.log(`  Transaction hash: ${tx.hash}`);
    console.log(`  Waiting for confirmation...`);
    
    const receipt = await tx.wait();
    
    console.log(`  ✅ Transfer successful!`);
    console.log(`  Gas used: ${receipt.gasUsed.toString()}`);
    console.log("");
  } catch (error) {
    console.error(`❌ Error: ${error.message}`);
  }
}

async function mint(token, provider, to, amount) {
  const privateKey = process.env.PRIVATE_KEY;
  if (!privateKey) {
    console.error("❌ Error: PRIVATE_KEY environment variable not set");
    return;
  }

  const wallet = new ethers.Wallet(privateKey, provider);
  const tokenWithSigner = token.connect(wallet);

  console.log(`\n🪙  Minting ${amount} GOB to ${to}\n`);

  try {
    const amountWei = ethers.parseEther(amount);
    const tx = await tokenWithSigner.mint(to, amountWei);
    
    console.log(`  Transaction hash: ${tx.hash}`);
    console.log(`  Waiting for confirmation...`);
    
    const receipt = await tx.wait();
    
    console.log(`  ✅ Mint successful!`);
    console.log(`  Gas used: ${receipt.gasUsed.toString()}`);
    
    // Show updated supply
    const circulatingSupply = await token.circulatingSupply();
    const remainingMintable = await token.remainingMintable();
    console.log(`  New Circulating Supply: ${ethers.formatEther(circulatingSupply)} GOB`);
    console.log(`  Remaining Mintable: ${ethers.formatEther(remainingMintable)} GOB`);
    console.log("");
  } catch (error) {
    console.error(`❌ Error: ${error.message}`);
  }
}

async function burn(token, provider, amount) {
  const privateKey = process.env.PRIVATE_KEY;
  if (!privateKey) {
    console.error("❌ Error: PRIVATE_KEY environment variable not set");
    return;
  }

  const wallet = new ethers.Wallet(privateKey, provider);
  const tokenWithSigner = token.connect(wallet);

  console.log(`\n🔥 Burning ${amount} GOB\n`);

  try {
    const amountWei = ethers.parseEther(amount);
    const tx = await tokenWithSigner.burn(amountWei);
    
    console.log(`  Transaction hash: ${tx.hash}`);
    console.log(`  Waiting for confirmation...`);
    
    const receipt = await tx.wait();
    
    console.log(`  ✅ Burn successful!`);
    console.log(`  Gas used: ${receipt.gasUsed.toString()}`);
    
    // Show updated supply
    const circulatingSupply = await token.circulatingSupply();
    const remainingMintable = await token.remainingMintable();
    console.log(`  New Circulating Supply: ${ethers.formatEther(circulatingSupply)} GOB`);
    console.log(`  Remaining Mintable: ${ethers.formatEther(remainingMintable)} GOB`);
    console.log("");
  } catch (error) {
    console.error(`❌ Error: ${error.message}`);
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

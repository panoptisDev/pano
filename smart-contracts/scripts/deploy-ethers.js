import { ethers } from "ethers";
import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { dirname, join } from "path";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

async function main() {
  console.log("🚀 Deploying Green Goblin Token (GOB) to Pano Blockchain...\n");

  // Connect to Pano blockchain
  const provider = new ethers.JsonRpcProvider("http://127.0.0.1:9545");
  
  // Create wallet from private key (User 4 - has ~87,160 PANO)
  const privateKey = "0x5d7a1a73da20b4273a3071411e61e43f46ea8e3cc61f892f72c3bb3b283762da";
  const wallet = new ethers.Wallet(privateKey, provider);
  
  console.log("Deploying from account:", wallet.address);
  
  // Get deployer balance
  const balance = await provider.getBalance(wallet.address);
  console.log("Account balance:", ethers.formatEther(balance), "PANO\n");

  // Load contract artifact
  const artifactPath = join(__dirname, "../artifacts/contracts/GreenGoblinToken.sol/GreenGoblinToken.json");
  const artifact = JSON.parse(readFileSync(artifactPath, "utf8"));
  
  // Create contract factory
  const factory = new ethers.ContractFactory(artifact.abi, artifact.bytecode, wallet);
  
  // Constructor parameters
  const name = "Green Goblin";
  const symbol = "GOB";
  const initialOwner = wallet.address;
  
  console.log("Deploying GreenGoblinToken contract...");
  console.log("  Name:", name);
  console.log("  Symbol:", symbol);
  console.log("  Initial Owner:", initialOwner);
  console.log("\nWaiting for deployment transaction...");
  
  const token = await factory.deploy(name, symbol, initialOwner);
  
  console.log("Transaction hash:", token.deploymentTransaction().hash);
  console.log("Waiting for confirmation...");
  
  await token.waitForDeployment();
  
  const tokenAddress = await token.getAddress();
  
  console.log("\n✅ Green Goblin Token deployed successfully!");
  console.log("\n📍 Contract Address:", tokenAddress);
  
  // Get token details
  const totalSupply = await token.totalSupply();
  const circulatingSupply = await token.circulatingSupply();
  const mintCap = await token.MINT_CAP();
  const remainingMintable = await token.remainingMintable();
  const decimals = await token.decimals();
  
  console.log("\n📊 Token Details:");
  console.log("  Name:", await token.name());
  console.log("  Symbol:", await token.symbol());
  console.log("  Decimals:", decimals);
  console.log("  Total Supply:", ethers.formatUnits(totalSupply, decimals), "GOB");
  console.log("  Circulating Supply:", ethers.formatUnits(circulatingSupply, decimals), "GOB");
  console.log("  Mint Cap:", ethers.formatUnits(mintCap, decimals), "GOB");
  console.log("  Remaining Mintable:", ethers.formatUnits(remainingMintable, decimals), "GOB");
  console.log("  Owner:", await token.owner());
  
  // Get owner balance
  const ownerBalance = await token.balanceOf(wallet.address);
  console.log("\n💰 Owner Balance:", ethers.formatUnits(ownerBalance, decimals), "GOB");
  
  console.log("\n🎉 Deployment complete!");
  console.log("\n📝 Add this token to MetaMask:");
  console.log("  Token Contract Address:", tokenAddress);
  console.log("  Token Symbol: GOB");
  console.log("  Decimals:", decimals);
  
  // Save deployment info
  const deploymentInfo = {
    network: "pano-testnet-3",
    chainId: 4093,
    contractAddress: tokenAddress,
    transactionHash: token.deploymentTransaction().hash,
    deployer: wallet.address,
    name: "Green Goblin",
    symbol: "GOB",
    decimals: Number(decimals),
    initialSupply: ethers.formatUnits(totalSupply, decimals),
    mintCap: ethers.formatUnits(mintCap, decimals),
    deploymentTime: new Date().toISOString()
  };
  
  console.log("\n💾 Deployment Info:");
  console.log(JSON.stringify(deploymentInfo, null, 2));
  
  // Save to file
  import('fs').then(fs => {
    fs.writeFileSync(
      join(__dirname, "../deployment.json"),
      JSON.stringify(deploymentInfo, null, 2)
    );
    console.log("\n📄 Deployment info saved to deployment.json");
  });
}

main()
  .then(() => {
    console.log("\n✨ All done!");
    process.exit(0);
  })
  .catch((error) => {
    console.error("\n❌ Error:", error);
    process.exit(1);
  });

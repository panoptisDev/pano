import hre from "hardhat";
import { ethers } from "ethers";

async function main() {
  console.log("🚀 Deploying Green Goblin Token (GOB) to Pano Blockchain...\n");

  // Get the deployer account
  const [deployer] = await hre.ethers.getSigners();
  
  console.log("Deploying from account:", deployer.address);
  
  // Get deployer balance
  const balance = await hre.ethers.provider.getBalance(deployer.address);
  console.log("Account balance:", ethers.formatEther(balance), "PANO\n");

  // Deploy the contract
  console.log("Deploying GreenGoblinToken contract...");
  
  const GreenGoblinToken = await hre.ethers.getContractFactory("GreenGoblinToken");
  
  // Constructor parameters
  const name = "Green Goblin";
  const symbol = "GOB";
  const initialOwner = deployer.address;
  
  const token = await GreenGoblinToken.deploy(name, symbol, initialOwner);
  
  await token.waitForDeployment();
  
  const tokenAddress = await token.getAddress();
  
  console.log("✅ Green Goblin Token deployed successfully!");
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
  const ownerBalance = await token.balanceOf(deployer.address);
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
    deployer: deployer.address,
    name: "Green Goblin",
    symbol: "GOB",
    decimals: Number(decimals),
    initialSupply: ethers.formatUnits(totalSupply, decimals),
    mintCap: ethers.formatUnits(mintCap, decimals),
    deploymentTime: new Date().toISOString()
  };
  
  console.log("\n💾 Deployment Info:");
  console.log(JSON.stringify(deploymentInfo, null, 2));
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

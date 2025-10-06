#!/usr/bin/env node

const { ethers } = require('ethers');

// Connect to local node
const provider = new ethers.JsonRpcProvider('http://127.0.0.1:9545');

// User 4 account (sender)
const user4PrivateKey = '0x5d7a1a73da20b4273a3071411e61e43f46ea8e3cc61f892f72c3bb3b283762da';
const user4Wallet = new ethers.Wallet(user4PrivateKey, provider);

// User 5 address (recipient)
const user5Address = '0x5Ab49BdE3137bE3e1285319B5F789d9f2831d9B5';

async function testFeeStructure() {
    try {
        console.log('=== Testing Exact Fee Structure ===\n');
        
        // Get balances BEFORE transaction
        console.log('📊 BEFORE TRANSACTION:');
        const balanceUser4Before = await provider.getBalance(user4Wallet.address);
        const balanceUser5Before = await provider.getBalance(user5Address);
        
        console.log(`User 4 (${user4Wallet.address}):`);
        console.log(`  Balance: ${ethers.formatEther(balanceUser4Before)} PANO`);
        console.log(`  Wei: ${balanceUser4Before.toString()}`);
        
        console.log(`\nUser 5 (${user5Address}):`);
        console.log(`  Balance: ${ethers.formatEther(balanceUser5Before)} PANO`);
        console.log(`  Wei: ${balanceUser5Before.toString()}`);
        
        // Get current gas price
        const feeData = await provider.getFeeData();
        const gasPrice = feeData.gasPrice;
        
        console.log(`\n⛽ Gas Price: ${ethers.formatUnits(gasPrice, 'gwei')} Gwei (${gasPrice.toString()} wei)`);
        
        // Amount to send: 100 PANO (nice round number)
        const amount = ethers.parseEther('100');
        console.log(`\n💸 Sending: ${ethers.formatEther(amount)} PANO`);
        
        // Calculate expected gas cost
        const gasLimit = 21000n;
        const expectedGasCost = gasPrice * gasLimit;
        console.log(`\n🔧 Expected Gas Cost:`);
        console.log(`  Gas Limit: ${gasLimit.toString()}`);
        console.log(`  Gas Price: ${gasPrice.toString()} wei`);
        console.log(`  Total Gas Cost: ${ethers.formatEther(expectedGasCost)} PANO (${expectedGasCost.toString()} wei)`);
        
        // Create transaction
        const tx = {
            to: user5Address,
            value: amount,
            gasLimit: gasLimit,
            gasPrice: gasPrice,
        };
        
        console.log('\n📤 Sending transaction...');
        const txResponse = await user4Wallet.sendTransaction(tx);
        console.log(`Transaction hash: ${txResponse.hash}`);
        
        console.log('⏳ Waiting for confirmation...');
        const receipt = await txResponse.wait();
        
        console.log(`\n✅ Transaction confirmed in block ${receipt.blockNumber}`);
        console.log(`Gas used: ${receipt.gasUsed.toString()}`);
        console.log(`Effective gas price: ${receipt.gasPrice.toString()} wei`);
        
        // Calculate actual gas cost
        const actualGasCost = receipt.gasUsed * receipt.gasPrice;
        console.log(`Actual gas cost: ${ethers.formatEther(actualGasCost)} PANO (${actualGasCost.toString()} wei)`);
        
        // Get balances AFTER transaction
        console.log('\n📊 AFTER TRANSACTION:');
        const balanceUser4After = await provider.getBalance(user4Wallet.address);
        const balanceUser5After = await provider.getBalance(user5Address);
        
        console.log(`User 4 (${user4Wallet.address}):`);
        console.log(`  Balance: ${ethers.formatEther(balanceUser4After)} PANO`);
        console.log(`  Wei: ${balanceUser4After.toString()}`);
        
        console.log(`\nUser 5 (${user5Address}):`);
        console.log(`  Balance: ${ethers.formatEther(balanceUser5After)} PANO`);
        console.log(`  Wei: ${balanceUser5After.toString()}`);
        
        // Calculate changes
        console.log('\n📈 BALANCE CHANGES:');
        const user4Change = balanceUser4After - balanceUser4Before;
        const user5Change = balanceUser5After - balanceUser5Before;
        
        console.log(`User 4: ${ethers.formatEther(user4Change)} PANO (${user4Change.toString()} wei)`);
        console.log(`User 5: ${ethers.formatEther(user5Change)} PANO (${user5Change.toString()} wei)`);
        
        // Verify the math
        console.log('\n🔍 VERIFICATION:');
        const expectedUser4Change = -(amount + actualGasCost);
        const expectedUser5Change = amount;
        
        console.log(`Expected User 4 change: ${ethers.formatEther(expectedUser4Change)} PANO`);
        console.log(`Actual User 4 change:   ${ethers.formatEther(user4Change)} PANO`);
        console.log(`Match: ${expectedUser4Change === user4Change ? '✅' : '❌'}`);
        
        console.log(`\nExpected User 5 change: ${ethers.formatEther(expectedUser5Change)} PANO`);
        console.log(`Actual User 5 change:   ${ethers.formatEther(user5Change)} PANO`);
        console.log(`Match: ${expectedUser5Change === user5Change ? '✅' : '❌'}`);
        
        // Summary
        console.log('\n📋 SUMMARY:');
        console.log(`Amount sent: ${ethers.formatEther(amount)} PANO`);
        console.log(`Gas fee: ${ethers.formatEther(actualGasCost)} PANO`);
        console.log(`Total cost to sender: ${ethers.formatEther(amount + actualGasCost)} PANO`);
        console.log(`Gas fee as % of transfer: ${(Number(ethers.formatEther(actualGasCost)) / Number(ethers.formatEther(amount)) * 100).toFixed(6)}%`);
        
    } catch (error) {
        console.error('\n❌ Error:', error.message);
        if (error.data) {
            console.error('Error data:', error.data);
        }
    }
}

testFeeStructure();

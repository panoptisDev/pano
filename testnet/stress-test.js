#!/usr/bin/env node

const { ethers } = require('ethers');

// Connect to local node
const provider = new ethers.JsonRpcProvider('http://127.0.0.1:9545');

// User accounts (senders - have ~100K PANO each)
const senders = [
    {
        name: 'User 4',
        privateKey: '0x5d7a1a73da20b4273a3071411e61e43f46ea8e3cc61f892f72c3bb3b283762da',
        address: '0xE56E6757b8D4124B235436a246af5DCB0a69D14D'
    },
    {
        name: 'User 5',
        privateKey: '0x5152c0b669f29ae1911ef16a597097709d963b99b13ab5c3632881c893c8be4e',
        address: '0x5Ab49BdE3137bE3e1285319B5F789d9f2831d9B5'
    }
];

// Validator accounts (recipients - have 1B PANO each)
const recipients = [
    '0xBcA3d19C24a0ebFc02b4047977F9473C388e4E98', // Validator 1
    '0x993669a7793F24b5F2e81c03dB494e0a83EAAE17', // Validator 2
    '0x649A72A7c3b30a8a347dC7A549D3e50c3eD4c97c'  // Validator 3
];

// Varying amounts in PANO
const amounts = [
    '50',    // 50 PANO
    '100',   // 100 PANO
    '200',   // 200 PANO
    '500',   // 500 PANO
    '1000',  // 1000 PANO
    '150',   // 150 PANO
    '300',   // 300 PANO
    '750',   // 750 PANO
    '250',   // 250 PANO
    '400'    // 400 PANO
];

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function sendTransaction(sender, recipient, amount, txNum) {
    try {
        const wallet = new ethers.Wallet(sender.privateKey, provider);
        
        // Get current gas price from network
        const feeData = await provider.getFeeData();
        const gasPrice = feeData.gasPrice || ethers.parseUnits('10', 'gwei');
        
        const tx = {
            to: recipient,
            value: ethers.parseEther(amount),
            gasLimit: 21000,
            gasPrice: gasPrice,
        };
        
        console.log(`\n[TX ${txNum}] ${sender.name} → ${recipient.slice(0,10)}...`);
        console.log(`  Amount: ${amount} PANO`);
        console.log(`  Gas Price: ${ethers.formatUnits(gasPrice, 'gwei')} Gwei`);
        
        const txResponse = await wallet.sendTransaction(tx);
        console.log(`  Hash: ${txResponse.hash}`);
        console.log(`  Status: ✓ Sent`);
        
        return txResponse;
        
    } catch (error) {
        console.error(`  Error: ${error.message}`);
        return null;
    }
}

async function runStressTest() {
    console.log('=== PANO Blockchain Stress Test ===');
    console.log('Sending 10 transactions with 10-second delays');
    console.log('From: User 4 and User 5');
    console.log('To: Validator addresses\n');
    
    const startTime = Date.now();
    const transactions = [];
    
    for (let i = 0; i < 10; i++) {
        const senderIndex = i % senders.length;
        const recipientIndex = i % recipients.length;
        const sender = senders[senderIndex];
        const recipient = recipients[recipientIndex];
        const amount = amounts[i];
        
        console.log(`\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`);
        console.log(`Transaction ${i + 1}/10`);
        console.log(`Time: ${new Date().toLocaleTimeString()}`);
        
        const tx = await sendTransaction(sender, recipient, amount, i + 1);
        if (tx) {
            transactions.push({
                num: i + 1,
                hash: tx.hash,
                sender: sender.name,
                amount: amount
            });
        }
        
        // Wait 10 seconds before next transaction (except after last one)
        if (i < 9) {
            console.log(`\n⏳ Waiting 10 seconds before next transaction...`);
            await sleep(10000);
        }
    }
    
    const endTime = Date.now();
    const totalTime = ((endTime - startTime) / 1000).toFixed(1);
    
    console.log('\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    console.log('=== Test Complete ===');
    console.log(`Total time: ${totalTime} seconds`);
    console.log(`Transactions sent: ${transactions.length}/10`);
    console.log(`Total PANO sent: ${amounts.reduce((sum, amt) => sum + parseFloat(amt), 0)} PANO`);
    
    if (transactions.length > 0) {
        console.log('\n=== Transaction Summary ===');
        for (const tx of transactions) {
            console.log(`[${tx.num}] ${tx.sender}: ${tx.amount} PANO - ${tx.hash}`);
        }
        
        console.log('\n💡 Tip: Wait ~60 seconds for blocks to be produced, then check balances');
    }
}

// Run the stress test
runStressTest().catch(console.error);

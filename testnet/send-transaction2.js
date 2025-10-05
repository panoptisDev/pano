#!/usr/bin/env node

const { ethers } = require('ethers');

// Connect to local node
const provider = new ethers.JsonRpcProvider('http://127.0.0.1:9545');

// User 5 account (sender) - sending back to User 4
const privateKey = '0x5152c0b669f29ae1911ef16a597097709d963b99b13ab5c3632881c893c8be4e';
const wallet = new ethers.Wallet(privateKey, provider);

// User 4 address (recipient)
const recipientAddress = '0xE56E6757b8D4124B235436a246af5DCB0a69D14D';

// Amount to send: 500 PANO (in wei: 500 * 10^18)
const amount = ethers.parseEther('500');

async function sendTransaction() {
    try {
        console.log('=== Sending PANO Transaction (User 5 → User 4) ===');
        console.log('From:', wallet.address);
        console.log('To:', recipientAddress);
        console.log('Amount:', ethers.formatEther(amount), 'PANO');
        
        // Get current gas price
        const feeData = await provider.getFeeData();
        console.log('Gas Price:', feeData.gasPrice ? feeData.gasPrice.toString() : 'unknown');
        
        // Create transaction
        const tx = {
            to: recipientAddress,
            value: amount,
            gasLimit: 21000,
            gasPrice: feeData.gasPrice || ethers.parseUnits('1', 'gwei'),
        };
        
        console.log('\nSending transaction...');
        const txResponse = await wallet.sendTransaction(tx);
        console.log('✓ Transaction submitted!');
        console.log('Transaction hash:', txResponse.hash);
        console.log('\nNote: Transaction was sent successfully. Check validator logs to confirm inclusion.');
        
    } catch (error) {
        console.error('Error:', error.message);
    }
}

sendTransaction();

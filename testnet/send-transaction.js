#!/usr/bin/env node

const { ethers } = require('ethers');

// Connect to local node
const provider = new ethers.JsonRpcProvider('http://127.0.0.1:9545');

// User 4 account (sender)
const privateKey = '0x5d7a1a73da20b4273a3071411e61e43f46ea8e3cc61f892f72c3bb3b283762da';
const wallet = new ethers.Wallet(privateKey, provider);

// User 5 address (recipient)
const recipientAddress = '0x5Ab49BdE3137bE3e1285319B5F789d9f2831d9B5';

// Amount to send: 1000 PANO (in wei: 1000 * 10^18)
const amount = ethers.parseEther('1000');

async function sendTransaction() {
    try {
        console.log('=== Sending PANO Transaction ===');
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
        console.log('Transaction hash:', txResponse.hash);
        
        console.log('Waiting for confirmation...');
        const receipt = await txResponse.wait();
        console.log('✓ Transaction confirmed in block:', receipt.blockNumber);
        console.log('Gas used:', receipt.gasUsed.toString());
        console.log('Status:', receipt.status === 1 ? 'Success' : 'Failed');
        
    } catch (error) {
        console.error('Error:', error.message);
        if (error.data) {
            console.error('Error data:', error.data);
        }
    }
}

sendTransaction();

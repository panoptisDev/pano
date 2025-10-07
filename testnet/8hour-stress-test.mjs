#!/usr/bin/env node

/**
 * Pano Blockchain 8-Hour Stress Test
 * Sends 10 transactions every 10 minutes for 8 hours
 * Tests chain stability and performance over extended period
 */

import { ethers } from 'ethers';
import fs from 'fs';

// Configuration
const RPC_URL = 'http://127.0.0.1:9545';
const CHAIN_ID = 4093;
const INTERVAL_MINUTES = 10;
const DURATION_HOURS = 8;
const TX_PER_BATCH = 10;

// Test accounts from testnet (we have private keys for these)
const SENDERS = [
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

// Recipient addresses (validators and other accounts - proper checksums)
const RECIPIENTS = [
    '0xBcA3d19C24a0ebFc02b4047977F9473C388e4E98',  // Validator 1
    '0x993669a7793F24b5F2e81c03dB494e0a83EAAE17',  // Validator 2
    '0x649A72A7c3b30a8a347dC7A549D3e50c3eD4c97c',  // Validator 3
];

// Colors for console output
const colors = {
    reset: '\x1b[0m',
    green: '\x1b[32m',
    blue: '\x1b[34m',
    yellow: '\x1b[33m',
    red: '\x1b[31m',
    cyan: '\x1b[36m'
};

// Stats tracking
let stats = {
    startTime: Date.now(),
    totalTxSent: 0,
    totalTxSuccess: 0,
    totalTxFailed: 0,
    totalConfirmed: 0,
    batches: []
};

// Setup provider and create wallets for all senders
const provider = new ethers.JsonRpcProvider(RPC_URL);
const wallets = SENDERS.map(sender => new ethers.Wallet(sender.privateKey, provider));

const LOG_FILE = `8hour-stress-test-${new Date().toISOString().replace(/:/g, '-').split('.')[0]}.log`;
const STATS_FILE = '8hour-stress-stats.json';

// Logging function
function log(message, color = null) {
    const timestamp = new Date().toISOString();
    const colorCode = color ? colors[color] : '';
    const resetCode = color ? colors.reset : '';
    const logMessage = `[${timestamp}] ${message}`;
    
    console.log(`${colorCode}${logMessage}${resetCode}`);
    fs.appendFileSync(LOG_FILE, logMessage + '\n');
}

// Display header
function displayHeader() {
    console.log(`${colors.blue}========================================${colors.reset}`);
    console.log(`${colors.blue}   Pano 8-Hour Stress Test${colors.reset}`);
    console.log(`${colors.blue}========================================${colors.reset}`);
    console.log('');
    console.log(`${colors.green}Configuration:${colors.reset}`);
    console.log(`  RPC URL: ${RPC_URL}`);
    console.log(`  Chain ID: ${CHAIN_ID}`);
    console.log(`  Interval: ${INTERVAL_MINUTES} minutes`);
    console.log(`  Duration: ${DURATION_HOURS} hours`);
    console.log(`  Transactions per batch: ${TX_PER_BATCH}`);
    console.log(`  Senders: ${SENDERS.length} accounts (${SENDERS.map(s => s.name).join(', ')})`);
    console.log(`  Recipients: ${RECIPIENTS.length} accounts`);
    console.log(`  Log file: ${LOG_FILE}`);
    console.log('');
    console.log(`${colors.yellow}Press Ctrl+C to stop the test${colors.reset}`);
    console.log('');
}

// Send a batch of transactions
async function sendBatch(batchNumber) {
    const timestamp = new Date().toLocaleString();
    log(`\n[Batch #${batchNumber}] ${timestamp}`, 'blue');
    
    const batchStats = {
        batchNumber,
        timestamp,
        transactions: [],
        success: 0,
        failed: 0,
        confirmed: 0
    };
    
    try {
        // Send TX_PER_BATCH transactions, rotating through senders
        for (let i = 0; i < TX_PER_BATCH; i++) {
            try {
                // Select sender (round-robin through available wallets)
                const senderIdx = i % wallets.length;
                const wallet = wallets[senderIdx];
                const senderName = SENDERS[senderIdx].name;
                
                // Get nonce for this sender
                const nonce = await wallet.getNonce();
                
                // Select recipient (round-robin)
                const recipientIdx = (batchNumber * TX_PER_BATCH + i) % RECIPIENTS.length;
                const recipient = RECIPIENTS[recipientIdx];
                
                // Prepare transaction
                const tx = {
                    to: recipient,
                    value: ethers.parseEther('0.01'), // 0.01 PANO
                    gasLimit: 21000,
                    gasPrice: ethers.parseUnits('1', 'gwei')
                };
                
                // Send transaction
                const txResponse = await wallet.sendTransaction(tx);
                log(`  ✓ TX #${i + 1}: ${txResponse.hash} (${senderName} -> ${recipient.substring(0, 10)}... nonce: ${nonce})`);
                
                batchStats.transactions.push({
                    txHash: txResponse.hash,
                    from: wallet.address,
                    to: recipient,
                    nonce: nonce,
                    status: 'pending'
                });
                
                batchStats.success++;
                stats.totalTxSuccess++;
                stats.totalTxSent++;
                
                // Small delay between transactions
                await new Promise(resolve => setTimeout(resolve, 500));
                
            } catch (error) {
                log(`  ✗ TX #${i + 1}: Failed - ${error.message}`, 'red');
                batchStats.failed++;
                stats.totalTxFailed++;
                stats.totalTxSent++;
            }
        }
        
        log(`  Batch result: ${batchStats.success} success, ${batchStats.failed} failed`);
        
        // Wait a bit for transactions to be mined, then check receipts
        await new Promise(resolve => setTimeout(resolve, 3000));
        
        let confirmed = 0;
        for (const tx of batchStats.transactions) {
            try {
                const receipt = await provider.getTransactionReceipt(tx.txHash);
                if (receipt && receipt.status === 1) {
                    tx.status = 'confirmed';
                    confirmed++;
                } else if (receipt && receipt.status === 0) {
                    tx.status = 'failed';
                }
            } catch (error) {
                // Receipt not available yet
            }
        }
        
        if (confirmed > 0) {
            log(`  Confirmed: ${confirmed}/${batchStats.transactions.length} transactions`, 'cyan');
            batchStats.confirmed = confirmed;
            stats.totalConfirmed += confirmed;
        }
        
    } catch (error) {
        log(`  Error in batch: ${error.message}`, 'red');
    }
    
    stats.batches.push(batchStats);
    updateStatsFile();
}

// Update stats file
function updateStatsFile() {
    const elapsed = (Date.now() - stats.startTime) / 1000;
    const elapsedHours = (elapsed / 3600).toFixed(2);
    const successRate = stats.totalTxSent > 0 
        ? ((stats.totalTxSuccess / stats.totalTxSent) * 100).toFixed(2) 
        : 0;
    const confirmRate = stats.totalTxSent > 0
        ? ((stats.totalConfirmed / stats.totalTxSent) * 100).toFixed(2)
        : 0;
    
    const statsData = {
        test_start: new Date(stats.startTime).toLocaleString(),
        elapsed_hours: parseFloat(elapsedHours),
        total_transactions_sent: stats.totalTxSent,
        total_success: stats.totalTxSuccess,
        total_failed: stats.totalTxFailed,
        total_confirmed: stats.totalConfirmed,
        success_rate: `${successRate}%`,
        confirmation_rate: `${confirmRate}%`,
        batches_completed: stats.batches.length,
        last_update: new Date().toLocaleString(),
        recent_batches: stats.batches.slice(-5).map(b => ({
            batch: b.batchNumber,
            timestamp: b.timestamp,
            success: b.success,
            failed: b.failed,
            confirmed: b.confirmed
        }))
    };
    
    fs.writeFileSync(STATS_FILE, JSON.stringify(statsData, null, 2));
}

// Display summary
function displaySummary() {
    const elapsed = (Date.now() - stats.startTime) / 1000;
    const elapsedHours = (elapsed / 3600).toFixed(2);
    const successRate = stats.totalTxSent > 0 
        ? ((stats.totalTxSuccess / stats.totalTxSent) * 100).toFixed(2) 
        : 0;
    const confirmRate = stats.totalTxSent > 0
        ? ((stats.totalConfirmed / stats.totalTxSent) * 100).toFixed(2)
        : 0;
    
    console.log('');
    console.log(`${colors.blue}========================================${colors.reset}`);
    console.log(`${colors.blue}   8-Hour Stress Test Summary${colors.reset}`);
    console.log(`${colors.blue}========================================${colors.reset}`);
    console.log(`  Duration: ${elapsedHours} hours`);
    console.log(`  Total batches: ${stats.batches.length}`);
    console.log(`  Total transactions: ${stats.totalTxSent}`);
    console.log(`  ${colors.green}Success: ${stats.totalTxSuccess}${colors.reset}`);
    console.log(`  ${colors.cyan}Confirmed: ${stats.totalConfirmed}${colors.reset}`);
    console.log(`  ${colors.red}Failed: ${stats.totalTxFailed}${colors.reset}`);
    console.log(`  Success rate: ${successRate}%`);
    console.log(`  Confirmation rate: ${confirmRate}%`);
    console.log('');
    console.log(`  Log file: ${LOG_FILE}`);
    console.log(`  Stats file: ${STATS_FILE}`);
    console.log(`${colors.blue}========================================${colors.reset}`);
}

// Cleanup on exit
function cleanup() {
    console.log('');
    log('Stopping stress test...', 'yellow');
    updateStatsFile();
    displaySummary();
    process.exit(0);
}

process.on('SIGINT', cleanup);
process.on('SIGTERM', cleanup);

// Main function
async function main() {
    displayHeader();
    
    const totalBatches = Math.floor((DURATION_HOURS * 60) / INTERVAL_MINUTES);
    log(`Starting 8-hour stress test (${totalBatches} batches total)`, 'green');
    
    // Check initial connection and balances for all senders
    try {
        const network = await provider.getNetwork();
        log(`Connected to network: Chain ID ${network.chainId}`, 'green');
        
        log('\nSender balances:', 'green');
        let totalBalance = BigInt(0);
        for (let i = 0; i < wallets.length; i++) {
            const balance = await provider.getBalance(wallets[i].address);
            totalBalance += balance;
            log(`  ${SENDERS[i].name}: ${ethers.formatEther(balance)} PANO (${wallets[i].address})`, 'cyan');
        }
        
        log(`  Total available: ${ethers.formatEther(totalBalance)} PANO`, 'green');
        
        const estimatedCost = ethers.parseEther('0.01') * BigInt(TX_PER_BATCH * totalBatches);
        log(`Estimated total cost: ${ethers.formatEther(estimatedCost)} PANO`, 'yellow');
        
        if (totalBalance < estimatedCost) {
            log(`WARNING: Total balance may be insufficient for full test!`, 'yellow');
        }
    } catch (error) {
        log(`Error connecting to network: ${error.message}`, 'red');
        process.exit(1);
    }
    
    // Run batches
    for (let batch = 1; batch <= totalBatches; batch++) {
        await sendBatch(batch);
        
        if (batch < totalBatches) {
            const waitMinutes = INTERVAL_MINUTES;
            const nextBatchTime = new Date(Date.now() + waitMinutes * 60 * 1000);
            log(`\nWaiting ${waitMinutes} minutes until next batch (at ${nextBatchTime.toLocaleTimeString()})...`, 'yellow');
            await new Promise(resolve => setTimeout(resolve, waitMinutes * 60 * 1000));
        }
    }
    
    log('\n8-hour stress test completed!', 'green');
    updateStatsFile();
    displaySummary();
}

// Run the test
main().catch(error => {
    console.error(`${colors.red}Fatal error: ${error.message}${colors.reset}`);
    console.error(error.stack);
    process.exit(1);
});

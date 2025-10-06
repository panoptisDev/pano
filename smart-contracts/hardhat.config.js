/** @type import('hardhat/config').HardhatUserConfig */
export default {
  solidity: {
    version: "0.8.20",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  networks: {
    pano: {
      type: 'http',
      url: "http://127.0.0.1:9545",
      chainId: 4093,
      accounts: [
        // User 4 private key (has ~87,160 PANO)
        "0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"
      ]
    }
  }
};

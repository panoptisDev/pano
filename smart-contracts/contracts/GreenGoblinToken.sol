// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";

contract GreenGoblinToken is ERC20, ERC20Burnable, Ownable, ERC20Permit {
    uint256 private _circulatingSupply;
    uint256 public constant MINT_CAP = 500_000_000 * 10**18; // 500M tokens

    event CirculatingSupplyUpdated(uint256 newSupply);

    constructor(
        string memory name, 
        string memory symbol,
        address initialOwner
    ) 
        ERC20(name, symbol) 
        Ownable(initialOwner) 
        ERC20Permit(name) 
    {
        require(initialOwner != address(0), "Owner cannot be zero address");
        
        uint256 initialSupply = 300_000_000 * 10**decimals(); // 300M tokens
        require(initialSupply <= MINT_CAP, "Initial supply exceeds cap");
        
        _mint(initialOwner, initialSupply);
        _circulatingSupply = initialSupply;
        emit CirculatingSupplyUpdated(_circulatingSupply);
    }

    function mint(address to, uint256 amount) public onlyOwner {
        require(_circulatingSupply + amount <= MINT_CAP, "Mint cap exceeded");
        _mint(to, amount);
        unchecked {
            _circulatingSupply += amount;
        }
        emit CirculatingSupplyUpdated(_circulatingSupply);
    }

    function burn(uint256 amount) public override {
        super.burn(amount);
        unchecked {
            _circulatingSupply -= amount;
        }
        emit CirculatingSupplyUpdated(_circulatingSupply);
    }

    function burnFrom(address from, uint256 amount) public override {
        super.burnFrom(from, amount);
        unchecked {
            _circulatingSupply -= amount;
        }
        emit CirculatingSupplyUpdated(_circulatingSupply);
    }

    function circulatingSupply() public view returns (uint256) {
        return _circulatingSupply;
    }

    function remainingMintable() public view returns (uint256) {
        return MINT_CAP - _circulatingSupply;
    }

    function totalSupply() public view override returns (uint256) {
        return _circulatingSupply; // Override to reflect circulating supply (excludes burned tokens)
    }
}

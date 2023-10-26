// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.18;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract DemoToken is ERC20 {
    constructor(uint256 initialSupply) ERC20("Demo", "DMO") {
        _mint(msg.sender, initialSupply);
    }
}
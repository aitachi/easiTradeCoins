// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "../src/TokenFactory.sol";
import "../src/Airdrop.sol";
import "../src/Staking.sol";

contract DeployAll is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");

        vm.startBroadcast(deployerPrivateKey);

        // Deploy TokenFactory
        TokenFactory tokenFactory = new TokenFactory();
        console.log("TokenFactory deployed at:", address(tokenFactory));

        // Deploy Airdrop
        Airdrop airdrop = new Airdrop();
        console.log("Airdrop deployed at:", address(airdrop));

        // Deploy Staking
        Staking staking = new Staking();
        console.log("Staking deployed at:", address(staking));

        vm.stopBroadcast();
    }
}

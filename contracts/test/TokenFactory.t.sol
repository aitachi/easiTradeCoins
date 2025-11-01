// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/EasiToken.sol";
import "../src/TokenFactory.sol";

contract TokenFactoryTest is Test {
    TokenFactory public factory;
    address public owner;
    address public user;

    function setUp() public {
        owner = address(this);
        user = address(0x1);
        factory = new TokenFactory();

        vm.deal(user, 10 ether);
    }

    function testCreateToken() public {
        vm.prank(user);
        address tokenAddress = factory.createToken{value: 0.01 ether}(
            "Test Token",
            "TEST",
            1000000 * 10**18
        );

        assertTrue(tokenAddress != address(0));

        TokenFactory.TokenInfo memory info = factory.getTokenInfo(tokenAddress);
        assertEq(info.name, "Test Token");
        assertEq(info.symbol, "TEST");
        assertEq(info.creator, user);
    }

    function testFailCreateTokenInsufficientFee() public {
        vm.prank(user);
        factory.createToken{value: 0.001 ether}(
            "Test Token",
            "TEST",
            1000000 * 10**18
        );
    }

    function testUpdateCreationFee() public {
        factory.updateCreationFee(0.02 ether);
        assertEq(factory.creationFee(), 0.02 ether);
    }

    function testWithdrawFees() public {
        vm.prank(user);
        factory.createToken{value: 0.01 ether}(
            "Test Token",
            "TEST",
            1000000 * 10**18
        );

        uint256 balanceBefore = owner.balance;
        factory.withdrawFees(payable(owner));
        uint256 balanceAfter = owner.balance;

        assertEq(balanceAfter - balanceBefore, 0.01 ether);
    }
}

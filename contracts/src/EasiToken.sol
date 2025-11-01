// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title EasiToken
 * @dev Standard ERC20 token with burn, pause, and access control features
 */
contract EasiToken is ERC20, ERC20Burnable, ERC20Pausable, AccessControl, ReentrancyGuard {
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");

    uint256 public constant MAX_SUPPLY = 1_000_000_000 * 10**18; // 1 billion max supply

    // Auto-burn configuration
    uint256 public autoBurnRate = 10; // 0.1% (10/10000)
    uint256 public constant RATE_DENOMINATOR = 10000;
    bool public autoBurnEnabled = false;

    // Total burned tokens
    uint256 public totalBurned;

    // Events
    event AutoBurnConfigured(uint256 rate, bool enabled);
    event TokensBurned(address indexed from, uint256 amount);
    event TokensMinted(address indexed to, uint256 amount);

    constructor(
        string memory name,
        string memory symbol,
        uint256 initialSupply,
        address admin
    ) ERC20(name, symbol) {
        require(initialSupply <= MAX_SUPPLY, "Initial supply exceeds max supply");
        require(admin != address(0), "Admin cannot be zero address");

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(MINTER_ROLE, admin);
        _grantRole(PAUSER_ROLE, admin);
        _grantRole(BURNER_ROLE, admin);

        if (initialSupply > 0) {
            _mint(admin, initialSupply);
            emit TokensMinted(admin, initialSupply);
        }
    }

    /**
     * @dev Mint new tokens
     */
    function mint(address to, uint256 amount) external onlyRole(MINTER_ROLE) {
        require(totalSupply() + amount <= MAX_SUPPLY, "Exceeds max supply");
        _mint(to, amount);
        emit TokensMinted(to, amount);
    }

    /**
     * @dev Pause token transfers
     */
    function pause() external onlyRole(PAUSER_ROLE) {
        _pause();
    }

    /**
     * @dev Unpause token transfers
     */
    function unpause() external onlyRole(PAUSER_ROLE) {
        _unpause();
    }

    /**
     * @dev Configure auto-burn mechanism
     */
    function configureAutoBurn(uint256 rate, bool enabled) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(rate <= 1000, "Rate too high"); // Max 10%
        autoBurnRate = rate;
        autoBurnEnabled = enabled;
        emit AutoBurnConfigured(rate, enabled);
    }

    /**
     * @dev Burn tokens from any address (with BURNER_ROLE)
     */
    function burnFrom(address account, uint256 amount) public override onlyRole(BURNER_ROLE) {
        super.burnFrom(account, amount);
        totalBurned += amount;
        emit TokensBurned(account, amount);
    }

    /**
     * @dev Override transfer to implement auto-burn
     */
    function _transfer(address from, address to, uint256 amount) internal virtual override(ERC20, ERC20Pausable) {
        if (autoBurnEnabled && from != address(0) && to != address(0)) {
            uint256 burnAmount = (amount * autoBurnRate) / RATE_DENOMINATOR;
            if (burnAmount > 0) {
                super._transfer(from, address(0), burnAmount);
                totalBurned += burnAmount;
                emit TokensBurned(from, burnAmount);
                amount -= burnAmount;
            }
        }
        super._transfer(from, to, amount);
    }

    /**
     * @dev Required override for _update
     */
    function _update(address from, address to, uint256 value) internal virtual override(ERC20, ERC20Pausable) {
        super._update(from, to, value);
    }
}

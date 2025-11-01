// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./EasiToken.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title TokenFactory
 * @dev Factory contract for creating ERC20 tokens
 */
contract TokenFactory is Ownable, ReentrancyGuard {

    struct TokenInfo {
        address tokenAddress;
        string name;
        string symbol;
        uint256 totalSupply;
        address creator;
        uint256 createdAt;
    }

    // Fee for creating tokens (in wei)
    uint256 public creationFee = 0.01 ether;

    // Maximum initial supply limit (10 billion tokens)
    uint256 public constant MAX_INITIAL_SUPPLY = 10_000_000_000 * 10**18;

    // Mapping from token address to token info
    mapping(address => TokenInfo) public tokens;

    // Mapping to track used names and symbols (lowercase hash to prevent duplicates)
    mapping(bytes32 => bool) public usedNames;
    mapping(bytes32 => bool) public usedSymbols;

    // Array of all created tokens
    address[] public allTokens;

    // Mapping from creator to their tokens
    mapping(address => address[]) public creatorTokens;

    // Events
    event TokenCreated(
        address indexed tokenAddress,
        string name,
        string symbol,
        uint256 initialSupply,
        address indexed creator,
        uint256 timestamp
    );

    event CreationFeeUpdated(uint256 oldFee, uint256 newFee);
    event FeesWithdrawn(address indexed to, uint256 amount);

    constructor() Ownable(msg.sender) {}

    /**
     * @dev Create a new ERC20 token
     */
    function createToken(
        string memory name,
        string memory symbol,
        uint256 initialSupply
    ) external payable nonReentrant returns (address) {
        require(msg.value >= creationFee, "Insufficient fee");
        require(bytes(name).length > 0, "Name cannot be empty");
        require(bytes(symbol).length > 0, "Symbol cannot be empty");
        require(initialSupply > 0, "Initial supply must be positive");
        require(initialSupply <= MAX_INITIAL_SUPPLY, "Initial supply exceeds maximum");

        // Check for duplicate names and symbols (case-insensitive)
        bytes32 nameHash = keccak256(abi.encodePacked(_toLower(name)));
        bytes32 symbolHash = keccak256(abi.encodePacked(_toLower(symbol)));

        require(!usedNames[nameHash], "Token name already exists");
        require(!usedSymbols[symbolHash], "Token symbol already exists");

        // Mark name and symbol as used
        usedNames[nameHash] = true;
        usedSymbols[symbolHash] = true;

        // Deploy new token contract
        EasiToken newToken = new EasiToken(
            name,
            symbol,
            initialSupply,
            msg.sender
        );

        address tokenAddress = address(newToken);

        // Store token info
        TokenInfo memory tokenInfo = TokenInfo({
            tokenAddress: tokenAddress,
            name: name,
            symbol: symbol,
            totalSupply: initialSupply,
            creator: msg.sender,
            createdAt: block.timestamp
        });

        tokens[tokenAddress] = tokenInfo;
        allTokens.push(tokenAddress);
        creatorTokens[msg.sender].push(tokenAddress);

        emit TokenCreated(
            tokenAddress,
            name,
            symbol,
            initialSupply,
            msg.sender,
            block.timestamp
        );

        // Refund excess payment
        if (msg.value > creationFee) {
            payable(msg.sender).transfer(msg.value - creationFee);
        }

        return tokenAddress;
    }

    /**
     * @dev Update creation fee
     */
    function updateCreationFee(uint256 newFee) external onlyOwner {
        uint256 oldFee = creationFee;
        creationFee = newFee;
        emit CreationFeeUpdated(oldFee, newFee);
    }

    /**
     * @dev Withdraw accumulated fees
     */
    function withdrawFees(address payable to) external onlyOwner {
        require(to != address(0), "Invalid address");
        uint256 balance = address(this).balance;
        require(balance > 0, "No fees to withdraw");

        to.transfer(balance);
        emit FeesWithdrawn(to, balance);
    }

    /**
     * @dev Get all tokens created by a user
     */
    function getCreatorTokens(address creator) external view returns (address[] memory) {
        return creatorTokens[creator];
    }

    /**
     * @dev Get total number of created tokens
     */
    function getTotalTokens() external view returns (uint256) {
        return allTokens.length;
    }

    /**
     * @dev Get token info
     */
    function getTokenInfo(address tokenAddress) external view returns (TokenInfo memory) {
        return tokens[tokenAddress];
    }

    /**
     * @dev Convert string to lowercase for case-insensitive comparison
     */
    function _toLower(string memory str) internal pure returns (string memory) {
        bytes memory bStr = bytes(str);
        bytes memory bLower = new bytes(bStr.length);

        for (uint i = 0; i < bStr.length; i++) {
            // Convert uppercase A-Z to lowercase a-z
            if ((uint8(bStr[i]) >= 65) && (uint8(bStr[i]) <= 90)) {
                bLower[i] = bytes1(uint8(bStr[i]) + 32);
            } else {
                bLower[i] = bStr[i];
            }
        }

        return string(bLower);
    }
}

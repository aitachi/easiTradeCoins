// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

/**
 * @title Airdrop
 * @dev Contract for managing token airdrops with Merkle tree verification
 */
contract Airdrop is Ownable, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    struct AirdropCampaign {
        IERC20 token;
        uint256 totalAmount;
        uint256 claimedAmount;
        uint256 startTime;
        uint256 endTime;
        bytes32 merkleRoot;
        bool active;
        uint256 minClaimInterval; // Minimum time between claims (anti-sybil)
        uint256 maxClaimPerAddress; // Maximum claims per address across campaigns
    }

    // Campaign ID counter
    uint256 public campaignCounter;

    // Mapping from campaign ID to campaign info
    mapping(uint256 => AirdropCampaign) public campaigns;

    // Mapping from campaign ID to address to claimed status
    mapping(uint256 => mapping(address => bool)) public hasClaimed;

    // Anti-sybil: Track total claims per address
    mapping(address => uint256) public totalClaimsByAddress;

    // Anti-sybil: Track last claim timestamp per address
    mapping(address => uint256) public lastClaimTime;

    // Anti-sybil: Global maximum claims per address
    uint256 public globalMaxClaimsPerAddress = 10;

    // Events
    event CampaignCreated(
        uint256 indexed campaignId,
        address indexed token,
        uint256 totalAmount,
        uint256 startTime,
        uint256 endTime
    );

    event TokensClaimed(
        uint256 indexed campaignId,
        address indexed beneficiary,
        uint256 amount
    );

    event CampaignCancelled(uint256 indexed campaignId);

    constructor() Ownable(msg.sender) {}

    /**
     * @dev Create a new airdrop campaign
     * @param tokenAddress Token contract address
     * @param totalAmount Total amount to distribute
     * @param startTime Campaign start time
     * @param endTime Campaign end time
     * @param merkleRoot Merkle tree root for verification
     * @param minClaimInterval Minimum seconds between claims (0 = no limit)
     * @param maxClaimPerAddress Max claims per address for this campaign (0 = no limit)
     */
    function createCampaign(
        address tokenAddress,
        uint256 totalAmount,
        uint256 startTime,
        uint256 endTime,
        bytes32 merkleRoot,
        uint256 minClaimInterval,
        uint256 maxClaimPerAddress
    ) external onlyOwner returns (uint256) {
        require(tokenAddress != address(0), "Invalid token address");
        require(totalAmount > 0, "Invalid amount");
        require(startTime < endTime, "Invalid time range");
        require(endTime > block.timestamp, "End time must be in future");
        require(merkleRoot != bytes32(0), "Invalid merkle root");

        campaignCounter++;
        uint256 campaignId = campaignCounter;

        IERC20 token = IERC20(tokenAddress);
        token.safeTransferFrom(msg.sender, address(this), totalAmount);

        campaigns[campaignId] = AirdropCampaign({
            token: token,
            totalAmount: totalAmount,
            claimedAmount: 0,
            startTime: startTime,
            endTime: endTime,
            merkleRoot: merkleRoot,
            active: true,
            minClaimInterval: minClaimInterval,
            maxClaimPerAddress: maxClaimPerAddress
        });

        emit CampaignCreated(campaignId, tokenAddress, totalAmount, startTime, endTime);

        return campaignId;
    }

    /**
     * @dev Claim airdrop tokens
     * @param campaignId Campaign ID
     * @param amount Amount to claim
     * @param merkleProof Merkle proof for verification
     */
    function claim(
        uint256 campaignId,
        uint256 amount,
        bytes32[] calldata merkleProof
    ) external nonReentrant whenNotPaused {
        AirdropCampaign storage campaign = campaigns[campaignId];

        require(campaign.active, "Campaign not active");
        require(block.timestamp >= campaign.startTime, "Campaign not started");
        require(block.timestamp <= campaign.endTime, "Campaign ended");
        require(!hasClaimed[campaignId][msg.sender], "Already claimed");
        require(amount > 0, "Invalid amount");

        // Anti-sybil checks
        require(
            totalClaimsByAddress[msg.sender] < globalMaxClaimsPerAddress,
            "Exceeded global claim limit"
        );

        if (campaign.minClaimInterval > 0) {
            require(
                block.timestamp >= lastClaimTime[msg.sender] + campaign.minClaimInterval,
                "Claim too soon"
            );
        }

        if (campaign.maxClaimPerAddress > 0) {
            require(
                totalClaimsByAddress[msg.sender] < campaign.maxClaimPerAddress,
                "Exceeded campaign claim limit"
            );
        }

        // Verify merkle proof
        bytes32 leaf = keccak256(abi.encodePacked(msg.sender, amount));
        require(verifyMerkleProof(merkleProof, campaign.merkleRoot, leaf), "Invalid proof");

        // Update state
        hasClaimed[campaignId][msg.sender] = true;
        campaign.claimedAmount += amount;
        totalClaimsByAddress[msg.sender]++;
        lastClaimTime[msg.sender] = block.timestamp;

        campaign.token.safeTransfer(msg.sender, amount);

        emit TokensClaimed(campaignId, msg.sender, amount);
    }

    /**
     * @dev Cancel campaign and refund remaining tokens
     */
    function cancelCampaign(uint256 campaignId) external onlyOwner {
        AirdropCampaign storage campaign = campaigns[campaignId];
        require(campaign.active, "Campaign not active");

        campaign.active = false;

        uint256 remainingAmount = campaign.totalAmount - campaign.claimedAmount;
        if (remainingAmount > 0) {
            campaign.token.safeTransfer(owner(), remainingAmount);
        }

        emit CampaignCancelled(campaignId);
    }

    /**
     * @dev Verify merkle proof
     */
    function verifyMerkleProof(
        bytes32[] memory proof,
        bytes32 root,
        bytes32 leaf
    ) internal pure returns (bool) {
        bytes32 computedHash = leaf;

        for (uint256 i = 0; i < proof.length; i++) {
            bytes32 proofElement = proof[i];

            if (computedHash <= proofElement) {
                computedHash = keccak256(abi.encodePacked(computedHash, proofElement));
            } else {
                computedHash = keccak256(abi.encodePacked(proofElement, computedHash));
            }
        }

        return computedHash == root;
    }

    /**
     * @dev Pause contract
     */
    function pause() external onlyOwner {
        _pause();
    }

    /**
     * @dev Unpause contract
     */
    function unpause() external onlyOwner {
        _unpause();
    }

    /**
     * @dev Update global max claims per address
     */
    function setGlobalMaxClaimsPerAddress(uint256 newLimit) external onlyOwner {
        require(newLimit > 0, "Limit must be greater than 0");
        globalMaxClaimsPerAddress = newLimit;
    }

    /**
     * @dev Get campaign details
     */
    function getCampaign(uint256 campaignId) external view returns (AirdropCampaign memory) {
        return campaigns[campaignId];
    }

    /**
     * @dev Check if address has claimed
     */
    function hasAddressClaimed(uint256 campaignId, address user) external view returns (bool) {
        return hasClaimed[campaignId][user];
    }
}

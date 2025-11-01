// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title Staking
 * @dev Token staking contract with flexible lock periods and rewards
 */
contract Staking is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    struct StakingPool {
        IERC20 stakingToken;
        IERC20 rewardToken;
        uint256 rewardRate; // Reward per second
        uint256 lockPeriod; // Lock period in seconds
        uint256 totalStaked;
        uint256 lastUpdateTime;
        uint256 rewardPerTokenStored;
        bool active;
    }

    struct UserStake {
        uint256 amount;
        uint256 rewardPerTokenPaid;
        uint256 rewards;
        uint256 stakeTime;
        uint256 unlockTime;
    }

    // Pool ID counter
    uint256 public poolCounter;

    // Mapping from pool ID to pool info
    mapping(uint256 => StakingPool) public pools;

    // Mapping from pool ID to user address to stake info
    mapping(uint256 => mapping(address => UserStake)) public stakes;

    // Early withdrawal penalty (10% = 1000)
    uint256 public constant PENALTY_RATE = 1000;
    uint256 public constant RATE_DENOMINATOR = 10000;

    // Events
    event PoolCreated(
        uint256 indexed poolId,
        address indexed stakingToken,
        address indexed rewardToken,
        uint256 rewardRate,
        uint256 lockPeriod
    );

    event Staked(
        uint256 indexed poolId,
        address indexed user,
        uint256 amount,
        uint256 unlockTime
    );

    event Withdrawn(
        uint256 indexed poolId,
        address indexed user,
        uint256 amount,
        uint256 penalty
    );

    event RewardClaimed(
        uint256 indexed poolId,
        address indexed user,
        uint256 reward
    );

    constructor() Ownable(msg.sender) {}

    /**
     * @dev Create a new staking pool
     */
    function createPool(
        address stakingToken,
        address rewardToken,
        uint256 rewardRate,
        uint256 lockPeriod
    ) external onlyOwner returns (uint256) {
        require(stakingToken != address(0), "Invalid staking token");
        require(rewardToken != address(0), "Invalid reward token");
        require(rewardRate > 0, "Invalid reward rate");

        poolCounter++;
        uint256 poolId = poolCounter;

        pools[poolId] = StakingPool({
            stakingToken: IERC20(stakingToken),
            rewardToken: IERC20(rewardToken),
            rewardRate: rewardRate,
            lockPeriod: lockPeriod,
            totalStaked: 0,
            lastUpdateTime: block.timestamp,
            rewardPerTokenStored: 0,
            active: true
        });

        emit PoolCreated(poolId, stakingToken, rewardToken, rewardRate, lockPeriod);

        return poolId;
    }

    /**
     * @dev Stake tokens
     */
    function stake(uint256 poolId, uint256 amount) external nonReentrant {
        StakingPool storage pool = pools[poolId];
        require(pool.active, "Pool not active");
        require(amount > 0, "Cannot stake 0");

        updateReward(poolId, msg.sender);

        UserStake storage userStake = stakes[poolId][msg.sender];

        pool.stakingToken.safeTransferFrom(msg.sender, address(this), amount);

        userStake.amount += amount;
        userStake.stakeTime = block.timestamp;
        userStake.unlockTime = block.timestamp + pool.lockPeriod;
        pool.totalStaked += amount;

        emit Staked(poolId, msg.sender, amount, userStake.unlockTime);
    }

    /**
     * @dev Withdraw staked tokens
     */
    function withdraw(uint256 poolId, uint256 amount) external nonReentrant {
        StakingPool storage pool = pools[poolId];
        UserStake storage userStake = stakes[poolId][msg.sender];

        require(amount > 0, "Cannot withdraw 0");
        require(userStake.amount >= amount, "Insufficient stake");

        updateReward(poolId, msg.sender);

        uint256 penalty = 0;

        // Apply penalty for early withdrawal
        if (block.timestamp < userStake.unlockTime) {
            penalty = (amount * PENALTY_RATE) / RATE_DENOMINATOR;
            amount -= penalty;
        }

        userStake.amount -= (amount + penalty);
        pool.totalStaked -= (amount + penalty);

        pool.stakingToken.safeTransfer(msg.sender, amount);

        if (penalty > 0) {
            // Transfer penalty to owner
            pool.stakingToken.safeTransfer(owner(), penalty);
        }

        emit Withdrawn(poolId, msg.sender, amount, penalty);
    }

    /**
     * @dev Claim rewards
     */
    function claimReward(uint256 poolId) external nonReentrant {
        updateReward(poolId, msg.sender);

        UserStake storage userStake = stakes[poolId][msg.sender];
        uint256 reward = userStake.rewards;

        require(reward > 0, "No rewards to claim");

        userStake.rewards = 0;
        pools[poolId].rewardToken.safeTransfer(msg.sender, reward);

        emit RewardClaimed(poolId, msg.sender, reward);
    }

    /**
     * @dev Update reward for user
     */
    function updateReward(uint256 poolId, address account) internal {
        StakingPool storage pool = pools[poolId];
        pool.rewardPerTokenStored = rewardPerToken(poolId);
        pool.lastUpdateTime = block.timestamp;

        if (account != address(0)) {
            UserStake storage userStake = stakes[poolId][account];
            userStake.rewards = earned(poolId, account);
            userStake.rewardPerTokenPaid = pool.rewardPerTokenStored;
        }
    }

    /**
     * @dev Calculate reward per token
     */
    function rewardPerToken(uint256 poolId) public view returns (uint256) {
        StakingPool storage pool = pools[poolId];

        if (pool.totalStaked == 0) {
            return pool.rewardPerTokenStored;
        }

        return pool.rewardPerTokenStored +
            ((block.timestamp - pool.lastUpdateTime) * pool.rewardRate * 1e18) / pool.totalStaked;
    }

    /**
     * @dev Calculate earned rewards for user
     */
    function earned(uint256 poolId, address account) public view returns (uint256) {
        UserStake storage userStake = stakes[poolId][account];
        return (userStake.amount * (rewardPerToken(poolId) - userStake.rewardPerTokenPaid)) / 1e18
            + userStake.rewards;
    }

    /**
     * @dev Get user stake info
     */
    function getUserStake(uint256 poolId, address account) external view returns (UserStake memory) {
        return stakes[poolId][account];
    }

    /**
     * @dev Deactivate pool
     */
    function deactivatePool(uint256 poolId) external onlyOwner {
        pools[poolId].active = false;
    }
}

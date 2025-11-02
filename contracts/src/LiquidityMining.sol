// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title LiquidityMining
 * @dev Liquidity mining and yield farming contract
 */
contract LiquidityMining is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    struct PoolInfo {
        IERC20 lpToken;           // LP token address
        uint256 allocPoint;       // Allocation points for this pool
        uint256 lastRewardBlock;  // Last block number that rewards distribution occurred
        uint256 accRewardPerShare; // Accumulated rewards per share, scaled by 1e12
        uint256 totalStaked;      // Total amount of LP tokens staked
    }

    struct UserInfo {
        uint256 amount;           // Amount of LP tokens staked
        uint256 rewardDebt;       // Reward debt
        uint256 pendingRewards;   // Pending rewards to be claimed
    }

    // Reward token
    IERC20 public rewardToken;

    // Reward per block
    uint256 public rewardPerBlock;

    // Pool information
    PoolInfo[] public poolInfo;

    // User information: poolId => user => UserInfo
    mapping(uint256 => mapping(address => UserInfo)) public userInfo;

    // Total allocation points
    uint256 public totalAllocPoint = 0;

    // Start block
    uint256 public startBlock;

    // Events
    event Deposit(address indexed user, uint256 indexed pid, uint256 amount);
    event Withdraw(address indexed user, uint256 indexed pid, uint256 amount);
    event EmergencyWithdraw(address indexed user, uint256 indexed pid, uint256 amount);
    event RewardClaimed(address indexed user, uint256 indexed pid, uint256 amount);

    constructor(
        IERC20 _rewardToken,
        uint256 _rewardPerBlock,
        uint256 _startBlock
    ) {
        rewardToken = _rewardToken;
        rewardPerBlock = _rewardPerBlock;
        startBlock = _startBlock;
    }

    /**
     * @dev Add a new LP token pool
     */
    function addPool(
        uint256 _allocPoint,
        IERC20 _lpToken,
        bool _withUpdate
    ) external onlyOwner {
        if (_withUpdate) {
            massUpdatePools();
        }

        uint256 lastRewardBlock = block.number > startBlock ? block.number : startBlock;
        totalAllocPoint += _allocPoint;

        poolInfo.push(
            PoolInfo({
                lpToken: _lpToken,
                allocPoint: _allocPoint,
                lastRewardBlock: lastRewardBlock,
                accRewardPerShare: 0,
                totalStaked: 0
            })
        );
    }

    /**
     * @dev Update pool allocation points
     */
    function setPool(
        uint256 _pid,
        uint256 _allocPoint,
        bool _withUpdate
    ) external onlyOwner {
        if (_withUpdate) {
            massUpdatePools();
        }

        totalAllocPoint = totalAllocPoint - poolInfo[_pid].allocPoint + _allocPoint;
        poolInfo[_pid].allocPoint = _allocPoint;
    }

    /**
     * @dev Update reward per block
     */
    function setRewardPerBlock(uint256 _rewardPerBlock, bool _withUpdate) external onlyOwner {
        if (_withUpdate) {
            massUpdatePools();
        }
        rewardPerBlock = _rewardPerBlock;
    }

    /**
     * @dev Get pending rewards for a user
     */
    function pendingReward(uint256 _pid, address _user) external view returns (uint256) {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][_user];

        uint256 accRewardPerShare = pool.accRewardPerShare;
        uint256 lpSupply = pool.totalStaked;

        if (block.number > pool.lastRewardBlock && lpSupply != 0) {
            uint256 blocks = block.number - pool.lastRewardBlock;
            uint256 reward = (blocks * rewardPerBlock * pool.allocPoint) / totalAllocPoint;
            accRewardPerShare += (reward * 1e12) / lpSupply;
        }

        return (user.amount * accRewardPerShare) / 1e12 - user.rewardDebt + user.pendingRewards;
    }

    /**
     * @dev Update all pools
     */
    function massUpdatePools() public {
        uint256 length = poolInfo.length;
        for (uint256 pid = 0; pid < length; ++pid) {
            updatePool(pid);
        }
    }

    /**
     * @dev Update reward variables of the given pool
     */
    function updatePool(uint256 _pid) public {
        PoolInfo storage pool = poolInfo[_pid];

        if (block.number <= pool.lastRewardBlock) {
            return;
        }

        uint256 lpSupply = pool.totalStaked;

        if (lpSupply == 0) {
            pool.lastRewardBlock = block.number;
            return;
        }

        uint256 blocks = block.number - pool.lastRewardBlock;
        uint256 reward = (blocks * rewardPerBlock * pool.allocPoint) / totalAllocPoint;

        pool.accRewardPerShare += (reward * 1e12) / lpSupply;
        pool.lastRewardBlock = block.number;
    }

    /**
     * @dev Stake LP tokens
     */
    function deposit(uint256 _pid, uint256 _amount) external nonReentrant {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][msg.sender];

        updatePool(_pid);

        if (user.amount > 0) {
            uint256 pending = (user.amount * pool.accRewardPerShare) / 1e12 - user.rewardDebt;
            if (pending > 0) {
                user.pendingRewards += pending;
            }
        }

        if (_amount > 0) {
            pool.lpToken.safeTransferFrom(msg.sender, address(this), _amount);
            user.amount += _amount;
            pool.totalStaked += _amount;
        }

        user.rewardDebt = (user.amount * pool.accRewardPerShare) / 1e12;

        emit Deposit(msg.sender, _pid, _amount);
    }

    /**
     * @dev Unstake LP tokens
     */
    function withdraw(uint256 _pid, uint256 _amount) external nonReentrant {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][msg.sender];

        require(user.amount >= _amount, "Insufficient balance");

        updatePool(_pid);

        uint256 pending = (user.amount * pool.accRewardPerShare) / 1e12 - user.rewardDebt;
        if (pending > 0) {
            user.pendingRewards += pending;
        }

        if (_amount > 0) {
            user.amount -= _amount;
            pool.totalStaked -= _amount;
            pool.lpToken.safeTransfer(msg.sender, _amount);
        }

        user.rewardDebt = (user.amount * pool.accRewardPerShare) / 1e12;

        emit Withdraw(msg.sender, _pid, _amount);
    }

    /**
     * @dev Claim pending rewards
     */
    function claim(uint256 _pid) external nonReentrant {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][msg.sender];

        updatePool(_pid);

        uint256 pending = (user.amount * pool.accRewardPerShare) / 1e12 - user.rewardDebt;
        pending += user.pendingRewards;

        if (pending > 0) {
            user.pendingRewards = 0;
            safeRewardTransfer(msg.sender, pending);
            emit RewardClaimed(msg.sender, _pid, pending);
        }

        user.rewardDebt = (user.amount * pool.accRewardPerShare) / 1e12;
    }

    /**
     * @dev Emergency withdraw without caring about rewards
     */
    function emergencyWithdraw(uint256 _pid) external nonReentrant {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][msg.sender];

        uint256 amount = user.amount;
        user.amount = 0;
        user.rewardDebt = 0;
        user.pendingRewards = 0;

        pool.totalStaked -= amount;
        pool.lpToken.safeTransfer(msg.sender, amount);

        emit EmergencyWithdraw(msg.sender, _pid, amount);
    }

    /**
     * @dev Safe reward transfer
     */
    function safeRewardTransfer(address _to, uint256 _amount) internal {
        uint256 rewardBal = rewardToken.balanceOf(address(this));
        if (_amount > rewardBal) {
            rewardToken.safeTransfer(_to, rewardBal);
        } else {
            rewardToken.safeTransfer(_to, _amount);
        }
    }

    /**
     * @dev Get pool count
     */
    function poolLength() external view returns (uint256) {
        return poolInfo.length;
    }

    /**
     * @dev Get user info
     */
    function getUserInfo(uint256 _pid, address _user)
        external
        view
        returns (
            uint256 amount,
            uint256 rewardDebt,
            uint256 pendingRewards
        )
    {
        UserInfo storage user = userInfo[_pid][_user];
        return (user.amount, user.rewardDebt, user.pendingRewards);
    }
}

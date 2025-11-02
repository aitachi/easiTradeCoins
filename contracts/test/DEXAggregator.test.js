const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time, loadFixture } = require("@nomicfoundation/hardhat-network-helpers");

describe("DEXAggregator", function () {
  async function deployDEXAggregatorFixture() {
    const [owner, user1, user2] = await ethers.getSigners();

    // Deploy mock ERC20 tokens
    const MockERC20 = await ethers.getContractFactory("MockERC20");
    const tokenA = await MockERC20.deploy("Token A", "TKA", ethers.parseEther("1000000"));
    const tokenB = await MockERC20.deploy("Token B", "TKB", ethers.parseEther("1000000"));

    // Deploy DEX Aggregator
    const DEXAggregator = await ethers.getContractFactory("DEXAggregator");
    const dexAggregator = await DEXAggregator.deploy(owner.address, 10); // 0.1% platform fee

    return { dexAggregator, tokenA, tokenB, owner, user1, user2 };
  }

  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      const { dexAggregator, owner } = await loadFixture(deployDEXAggregatorFixture);
      expect(await dexAggregator.owner()).to.equal(owner.address);
    });

    it("Should set the correct platform fee", async function () {
      const { dexAggregator } = await loadFixture(deployDEXAggregatorFixture);
      expect(await dexAggregator.platformFee()).to.equal(10);
    });
  });

  describe("DEX Management", function () {
    it("Should add DEX router", async function () {
      const { dexAggregator, owner } = await loadFixture(deployDEXAggregatorFixture);

      const mockRouter = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"; // Uniswap V2 Router
      await expect(dexAggregator.addDEX("Uniswap", mockRouter))
        .to.emit(dexAggregator, "DEXAdded")
        .withArgs("Uniswap", mockRouter);
    });

    it("Should remove DEX router", async function () {
      const { dexAggregator } = await loadFixture(deployDEXAggregatorFixture);

      const mockRouter = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D";
      await dexAggregator.addDEX("Uniswap", mockRouter);

      await expect(dexAggregator.removeDEX(mockRouter))
        .to.emit(dexAggregator, "DEXRemoved")
        .withArgs(mockRouter);
    });

    it("Should not allow non-owner to add DEX", async function () {
      const { dexAggregator, user1 } = await loadFixture(deployDEXAggregatorFixture);

      const mockRouter = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D";
      await expect(
        dexAggregator.connect(user1).addDEX("Uniswap", mockRouter)
      ).to.be.revertedWith("Ownable: caller is not the owner");
    });
  });

  describe("Fee Management", function () {
    it("Should update platform fee", async function () {
      const { dexAggregator } = await loadFixture(deployDEXAggregatorFixture);

      await dexAggregator.setPlatformFee(20); // 0.2%
      expect(await dexAggregator.platformFee()).to.equal(20);
    });

    it("Should not allow fee > 1%", async function () {
      const { dexAggregator } = await loadFixture(deployDEXAggregatorFixture);

      await expect(
        dexAggregator.setPlatformFee(101)
      ).to.be.revertedWith("Fee too high");
    });

    it("Should collect fees to treasury", async function () {
      const { dexAggregator, owner, user1 } = await loadFixture(deployDEXAggregatorFixture);

      // Update fee treasury
      await dexAggregator.setFeeTreasury(user1.address);
      expect(await dexAggregator.feeTreasury()).to.equal(user1.address);
    });
  });
});

describe("LiquidityMining", function () {
  async function deployLiquidityMiningFixture() {
    const [owner, user1, user2] = await ethers.getSigners();

    // Deploy reward token
    const MockERC20 = await ethers.getContractFactory("MockERC20");
    const rewardToken = await MockERC20.deploy("Reward Token", "RWD", ethers.parseEther("1000000"));
    const lpToken = await MockERC20.deploy("LP Token", "LP", ethers.parseEther("1000000"));

    // Deploy Liquidity Mining
    const currentBlock = await ethers.provider.getBlockNumber();
    const rewardPerBlock = ethers.parseEther("10"); // 10 tokens per block

    const LiquidityMining = await ethers.getContractFactory("LiquidityMining");
    const liquidityMining = await LiquidityMining.deploy(
      await rewardToken.getAddress(),
      rewardPerBlock,
      currentBlock + 10
    );

    // Transfer reward tokens to contract
    await rewardToken.transfer(await liquidityMining.getAddress(), ethers.parseEther("100000"));

    // Transfer LP tokens to users for testing
    await lpToken.transfer(user1.address, ethers.parseEther("1000"));
    await lpToken.transfer(user2.address, ethers.parseEther("1000"));

    return { liquidityMining, rewardToken, lpToken, owner, user1, user2 };
  }

  describe("Deployment", function () {
    it("Should set correct reward token", async function () {
      const { liquidityMining, rewardToken } = await loadFixture(deployLiquidityMiningFixture);
      expect(await liquidityMining.rewardToken()).to.equal(await rewardToken.getAddress());
    });

    it("Should set correct reward per block", async function () {
      const { liquidityMining } = await loadFixture(deployLiquidityMiningFixture);
      expect(await liquidityMining.rewardPerBlock()).to.equal(ethers.parseEther("10"));
    });
  });

  describe("Pool Management", function () {
    it("Should add new pool", async function () {
      const { liquidityMining, lpToken } = await loadFixture(deployLiquidityMiningFixture);

      await expect(liquidityMining.addPool(100, await lpToken.getAddress(), false))
        .to.emit(liquidityMining, "PoolAdded");

      const poolInfo = await liquidityMining.poolInfo(0);
      expect(poolInfo.allocPoint).to.equal(100);
    });

    it("Should update pool allocation", async function () {
      const { liquidityMining, lpToken } = await loadFixture(deployLiquidityMiningFixture);

      await liquidityMining.addPool(100, await lpToken.getAddress(), false);
      await liquidityMining.setPool(0, 200, false);

      const poolInfo = await liquidityMining.poolInfo(0);
      expect(poolInfo.allocPoint).to.equal(200);
    });
  });

  describe("Staking", function () {
    it("Should allow user to stake LP tokens", async function () {
      const { liquidityMining, lpToken, user1 } = await loadFixture(deployLiquidityMiningFixture);

      // Add pool
      await liquidityMining.addPool(100, await lpToken.getAddress(), false);

      // Approve and deposit
      const depositAmount = ethers.parseEther("100");
      await lpToken.connect(user1).approve(await liquidityMining.getAddress(), depositAmount);

      await expect(liquidityMining.connect(user1).deposit(0, depositAmount))
        .to.emit(liquidityMining, "Deposit")
        .withArgs(user1.address, 0, depositAmount);

      const userInfo = await liquidityMining.userInfo(0, user1.address);
      expect(userInfo.amount).to.equal(depositAmount);
    });

    it("Should allow user to withdraw LP tokens", async function () {
      const { liquidityMining, lpToken, user1 } = await loadFixture(deployLiquidityMiningFixture);

      await liquidityMining.addPool(100, await lpToken.getAddress(), false);

      const depositAmount = ethers.parseEther("100");
      await lpToken.connect(user1).approve(await liquidityMining.getAddress(), depositAmount);
      await liquidityMining.connect(user1).deposit(0, depositAmount);

      // Withdraw
      const withdrawAmount = ethers.parseEther("50");
      await expect(liquidityMining.connect(user1).withdraw(0, withdrawAmount))
        .to.emit(liquidityMining, "Withdraw")
        .withArgs(user1.address, 0, withdrawAmount);

      const userInfo = await liquidityMining.userInfo(0, user1.address);
      expect(userInfo.amount).to.equal(ethers.parseEther("50"));
    });

    it("Should calculate pending rewards correctly", async function () {
      const { liquidityMining, lpToken, user1 } = await loadFixture(deployLiquidityMiningFixture);

      await liquidityMining.addPool(100, await lpToken.getAddress(), false);

      const depositAmount = ethers.parseEther("100");
      await lpToken.connect(user1).approve(await liquidityMining.getAddress(), depositAmount);
      await liquidityMining.connect(user1).deposit(0, depositAmount);

      // Mine some blocks
      await time.mine(10);

      const pending = await liquidityMining.pendingReward(0, user1.address);
      expect(pending).to.be.gt(0);
    });
  });

  describe("Reward Distribution", function () {
    it("Should distribute rewards to multiple users", async function () {
      const { liquidityMining, lpToken, rewardToken, user1, user2 } = await loadFixture(deployLiquidityMiningFixture);

      await liquidityMining.addPool(100, await lpToken.getAddress(), false);

      // User1 deposits 60%
      const deposit1 = ethers.parseEther("60");
      await lpToken.connect(user1).approve(await liquidityMining.getAddress(), deposit1);
      await liquidityMining.connect(user1).deposit(0, deposit1);

      // User2 deposits 40%
      const deposit2 = ethers.parseEther("40");
      await lpToken.connect(user2).approve(await liquidityMining.getAddress(), deposit2);
      await liquidityMining.connect(user2).deposit(0, deposit2);

      // Mine blocks
      await time.mine(10);

      // Check rewards are proportional
      const pending1 = await liquidityMining.pendingReward(0, user1.address);
      const pending2 = await liquidityMining.pendingReward(0, user2.address);

      // User1 should have ~60% of rewards, User2 ~40%
      const ratio = Number(pending1) / Number(pending2);
      expect(ratio).to.be.closeTo(1.5, 0.1); // 60/40 = 1.5
    });

    it("Should allow claiming rewards", async function () {
      const { liquidityMining, lpToken, rewardToken, user1 } = await loadFixture(deployLiquidityMiningFixture);

      await liquidityMining.addPool(100, await lpToken.getAddress(), false);

      const depositAmount = ethers.parseEther("100");
      await lpToken.connect(user1).approve(await liquidityMining.getAddress(), depositAmount);
      await liquidityMining.connect(user1).deposit(0, depositAmount);

      // Mine blocks
      await time.mine(10);

      const balanceBefore = await rewardToken.balanceOf(user1.address);

      await expect(liquidityMining.connect(user1).claim(0))
        .to.emit(liquidityMining, "RewardClaimed");

      const balanceAfter = await rewardToken.balanceOf(user1.address);
      expect(balanceAfter).to.be.gt(balanceBefore);
    });
  });

  describe("Emergency Withdraw", function () {
    it("Should allow emergency withdraw without rewards", async function () {
      const { liquidityMining, lpToken, rewardToken, user1 } = await loadFixture(deployLiquidityMiningFixture);

      await liquidityMining.addPool(100, await lpToken.getAddress(), false);

      const depositAmount = ethers.parseEther("100");
      await lpToken.connect(user1).approve(await liquidityMining.getAddress(), depositAmount);
      await liquidityMining.connect(user1).deposit(0, depositAmount);

      const lpBalanceBefore = await lpToken.balanceOf(user1.address);
      const rewardBalanceBefore = await rewardToken.balanceOf(user1.address);

      await expect(liquidityMining.connect(user1).emergencyWithdraw(0))
        .to.emit(liquidityMining, "EmergencyWithdraw");

      const lpBalanceAfter = await lpToken.balanceOf(user1.address);
      const rewardBalanceAfter = await rewardToken.balanceOf(user1.address);

      // LP tokens returned
      expect(lpBalanceAfter).to.equal(lpBalanceBefore + depositAmount);
      // No rewards given
      expect(rewardBalanceAfter).to.equal(rewardBalanceBefore);
    });
  });
});

// Mock ERC20 contract for testing
describe("MockERC20 Helper", function () {
  it("Should be deployable", async function () {
    const MockERC20 = await ethers.getContractFactory("MockERC20");
    const token = await MockERC20.deploy("Test Token", "TEST", ethers.parseEther("1000"));

    expect(await token.name()).to.equal("Test Token");
    expect(await token.symbol()).to.equal("TEST");
  });
});

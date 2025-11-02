//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

// Sepolia Testnet Deployment and Testing Script
const hre = require("hardhat");
const fs = require("fs");

async function main() {
    console.log("====================================");
    console.log("Sepolia Testnet Deployment & Testing");
    console.log("====================================\n");

    const [deployer] = await ethers.getSigners();
    console.log("Deploying contracts with account:", deployer.address);
    console.log("Account balance:", ethers.formatEther(await ethers.provider.getBalance(deployer.address)), "ETH\n");

    const deploymentResults = {
        network: "sepolia",
        chainId: 11155111,
        deployer: deployer.address,
        timestamp: new Date().toISOString(),
        contracts: {}
    };

    // ================================
    // 1. Deploy Mock ERC20 Tokens (for testing)
    // ================================
    console.log("1. Deploying Mock Tokens...");
    console.log("================================");

    const MockERC20 = await ethers.getContractFactory("MockERC20");

    const tokenUSDT = await MockERC20.deploy(
        "Test USDT",
        "USDT",
        ethers.parseEther("10000000") // 10M tokens
    );
    await tokenUSDT.waitForDeployment();
    const usdtAddress = await tokenUSDT.getAddress();
    console.log("✓ USDT deployed to:", usdtAddress);

    const tokenBTC = await MockERC20.deploy(
        "Test BTC",
        "BTC",
        ethers.parseEther("21000") // 21K tokens
    );
    await tokenBTC.waitForDeployment();
    const btcAddress = await tokenBTC.getAddress();
    console.log("✓ BTC deployed to:", btcAddress);

    const tokenETH = await MockERC20.deploy(
        "Test ETH",
        "ETH",
        ethers.parseEther("1000000") // 1M tokens
    );
    await tokenETH.waitForDeployment();
    const ethAddress = await tokenETH.getAddress();
    console.log("✓ ETH deployed to:", ethAddress);

    deploymentResults.contracts.mockTokens = {
        USDT: usdtAddress,
        BTC: btcAddress,
        ETH: ethAddress
    };

    console.log("");

    // ================================
    // 2. Deploy DEX Aggregator
    // ================================
    console.log("2. Deploying DEX Aggregator...");
    console.log("================================");

    const DEXAggregator = await ethers.getContractFactory("DEXAggregator");
    const dexAggregator = await DEXAggregator.deploy(
        deployer.address, // fee treasury
        10 // 0.1% platform fee
    );
    await dexAggregator.waitForDeployment();
    const dexAggregatorAddress = await dexAggregator.getAddress();
    console.log("✓ DEXAggregator deployed to:", dexAggregatorAddress);

    deploymentResults.contracts.dexAggregator = {
        address: dexAggregatorAddress,
        feeTreasury: deployer.address,
        platformFee: 10
    };

    console.log("");

    // ================================
    // 3. Deploy Liquidity Mining
    // ================================
    console.log("3. Deploying Liquidity Mining...");
    console.log("================================");

    // Deploy reward token
    const rewardToken = await MockERC20.deploy(
        "EasiTrade Reward",
        "ETR",
        ethers.parseEther("100000000") // 100M tokens
    );
    await rewardToken.waitForDeployment();
    const rewardTokenAddress = await rewardToken.getAddress();
    console.log("✓ Reward Token deployed to:", rewardTokenAddress);

    // Get current block number
    const currentBlock = await ethers.provider.getBlockNumber();
    const startBlock = currentBlock + 100; // Start in 100 blocks
    const rewardPerBlock = ethers.parseEther("10"); // 10 tokens per block

    const LiquidityMining = await ethers.getContractFactory("LiquidityMining");
    const liquidityMining = await LiquidityMining.deploy(
        rewardTokenAddress,
        rewardPerBlock,
        startBlock
    );
    await liquidityMining.waitForDeployment();
    const liquidityMiningAddress = await liquidityMining.getAddress();
    console.log("✓ LiquidityMining deployed to:", liquidityMiningAddress);

    // Transfer reward tokens to mining contract
    const rewardAmount = ethers.parseEther("10000000"); // 10M tokens
    await rewardToken.transfer(liquidityMiningAddress, rewardAmount);
    console.log("✓ Transferred", ethers.formatEther(rewardAmount), "reward tokens to LiquidityMining");

    deploymentResults.contracts.liquidityMining = {
        address: liquidityMiningAddress,
        rewardToken: rewardTokenAddress,
        rewardPerBlock: ethers.formatEther(rewardPerBlock),
        startBlock: startBlock
    };

    console.log("");

    // ================================
    // 4. Setup Liquidity Mining Pools
    // ================================
    console.log("4. Setting up Mining Pools...");
    console.log("================================");

    // Create LP tokens for testing
    const lpTokenBTCUSDT = await MockERC20.deploy("BTC-USDT LP", "BTC-USDT-LP", ethers.parseEther("1000000"));
    await lpTokenBTCUSDT.waitForDeployment();
    const lpBTCUSDTAddress = await lpTokenBTCUSDT.getAddress();

    const lpTokenETHUSDT = await MockERC20.deploy("ETH-USDT LP", "ETH-USDT-LP", ethers.parseEther("1000000"));
    await lpTokenETHUSDT.waitForDeployment();
    const lpETHUSDTAddress = await lpTokenETHUSDT.getAddress();

    // Add pools
    await liquidityMining.addPool(100, lpBTCUSDTAddress, false); // 100 allocation points
    console.log("✓ Added BTC-USDT pool with 100 allocation points");

    await liquidityMining.addPool(80, lpETHUSDTAddress, false); // 80 allocation points
    console.log("✓ Added ETH-USDT pool with 80 allocation points");

    deploymentResults.contracts.lpTokens = {
        "BTC-USDT": lpBTCUSDTAddress,
        "ETH-USDT": lpETHUSDTAddress
    };

    console.log("");

    // ================================
    // 5. Contract Verification
    // ================================
    console.log("5. Running Contract Tests...");
    console.log("================================");

    try {
        // Test DEX Aggregator
        const platformFee = await dexAggregator.platformFee();
        console.log("✓ DEX Aggregator platform fee:", platformFee.toString());

        // Test Liquidity Mining
        const poolLength = await liquidityMining.poolLength();
        console.log("✓ Liquidity Mining pools:", poolLength.toString());

        const pool0 = await liquidityMining.poolInfo(0);
        console.log("✓ Pool 0 allocation points:", pool0.allocPoint.toString());

        console.log("\n✓ All contract tests passed!");
    } catch (error) {
        console.log("\n✗ Contract tests failed:", error.message);
    }

    console.log("");

    // ================================
    // 6. Save Deployment Info
    // ================================
    console.log("6. Saving Deployment Information...");
    console.log("================================");

    const deploymentFile = `./deployments/sepolia-${Date.now()}.json`;
    fs.writeFileSync(deploymentFile, JSON.stringify(deploymentResults, null, 2));
    console.log("✓ Deployment info saved to:", deploymentFile);

    console.log("");

    // ================================
    // 7. Deployment Summary
    // ================================
    console.log("====================================");
    console.log("Deployment Summary");
    console.log("====================================");
    console.log("Network: Sepolia Testnet");
    console.log("Chain ID:", deploymentResults.chainId);
    console.log("Deployer:", deploymentResults.deployer);
    console.log("");
    console.log("Deployed Contracts:");
    console.log("-----------------------------------");
    console.log("DEX Aggregator:", dexAggregatorAddress);
    console.log("Liquidity Mining:", liquidityMiningAddress);
    console.log("Reward Token:", rewardTokenAddress);
    console.log("USDT Token:", usdtAddress);
    console.log("BTC Token:", btcAddress);
    console.log("ETH Token:", ethAddress);
    console.log("");
    console.log("LP Tokens:");
    console.log("-----------------------------------");
    console.log("BTC-USDT LP:", lpBTCUSDTAddress);
    console.log("ETH-USDT LP:", lpETHUSDTAddress);
    console.log("");

    // ================================
    // 8. Etherscan Verification
    // ================================
    console.log("====================================");
    console.log("Verify contracts on Etherscan:");
    console.log("====================================");
    console.log(`npx hardhat verify --network sepolia ${dexAggregatorAddress} "${deployer.address}" 10`);
    console.log(`npx hardhat verify --network sepolia ${liquidityMiningAddress} "${rewardTokenAddress}" "${rewardPerBlock}" ${startBlock}`);
    console.log("");

    // ================================
    // 9. Next Steps
    // ================================
    console.log("====================================");
    console.log("Next Steps:");
    console.log("====================================");
    console.log("1. Verify contracts on Etherscan using the commands above");
    console.log("2. Update backend .env with contract addresses:");
    console.log(`   CONTRACT_ADDRESS_DEX_AGGREGATOR=${dexAggregatorAddress}`);
    console.log(`   CONTRACT_ADDRESS_LIQUIDITY_MINING=${liquidityMiningAddress}`);
    console.log("3. Test contract interactions via backend API");
    console.log("4. Monitor transactions on Sepolia Etherscan");
    console.log("");
    console.log("Sepolia Etherscan: https://sepolia.etherscan.io/");
    console.log("Sepolia Faucet: https://sepoliafaucet.com/");
    console.log("");
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });

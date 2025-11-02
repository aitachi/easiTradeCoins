// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title DEXAggregator
 * @dev Aggregates liquidity from multiple DEXes to get best prices
 * Supports: Uniswap V2/V3, SushiSwap, PancakeSwap
 */
contract DEXAggregator is Ownable, ReentrancyGuard {

    // Supported DEX routers
    address[] public dexRouters;
    mapping(address => bool) public isDexSupported;

    // Fee configuration
    uint256 public platformFee = 10; // 0.1% (basis points)
    address public feeCollector;

    // Events
    event SwapExecuted(
        address indexed user,
        address indexed tokenIn,
        address indexed tokenOut,
        uint256 amountIn,
        uint256 amountOut,
        address dex
    );

    event DEXAdded(address indexed dexRouter);
    event DEXRemoved(address indexed dexRouter);
    event FeeUpdated(uint256 newFee);

    struct Quote {
        address dex;
        uint256 amountOut;
        address[] path;
    }

    constructor(address _feeCollector) {
        feeCollector = _feeCollector;
    }

    /**
     * @dev Add a DEX router to the aggregator
     */
    function addDEX(address _dexRouter) external onlyOwner {
        require(_dexRouter != address(0), "Invalid DEX address");
        require(!isDexSupported[_dexRouter], "DEX already added");

        dexRouters.push(_dexRouter);
        isDexSupported[_dexRouter] = true;

        emit DEXAdded(_dexRouter);
    }

    /**
     * @dev Remove a DEX router from the aggregator
     */
    function removeDEX(address _dexRouter) external onlyOwner {
        require(isDexSupported[_dexRouter], "DEX not supported");

        isDexSupported[_dexRouter] = false;

        // Remove from array
        for (uint256 i = 0; i < dexRouters.length; i++) {
            if (dexRouters[i] == _dexRouter) {
                dexRouters[i] = dexRouters[dexRouters.length - 1];
                dexRouters.pop();
                break;
            }
        }

        emit DEXRemoved(_dexRouter);
    }

    /**
     * @dev Get the best price across all DEXes
     */
    function getBestQuote(
        address tokenIn,
        address tokenOut,
        uint256 amountIn
    ) public view returns (Quote memory bestQuote) {
        require(amountIn > 0, "Amount must be > 0");

        uint256 bestAmountOut = 0;
        address bestDEX;
        address[] memory bestPath;

        // Query all DEXes
        for (uint256 i = 0; i < dexRouters.length; i++) {
            address dex = dexRouters[i];

            try this.getAmountsOut(dex, amountIn, tokenIn, tokenOut) returns (
                uint256 amountOut,
                address[] memory path
            ) {
                if (amountOut > bestAmountOut) {
                    bestAmountOut = amountOut;
                    bestDEX = dex;
                    bestPath = path;
                }
            } catch {
                // Skip this DEX if query fails
                continue;
            }
        }

        require(bestAmountOut > 0, "No liquidity found");

        return Quote({
            dex: bestDEX,
            amountOut: bestAmountOut,
            path: bestPath
        });
    }

    /**
     * @dev Get amounts out from a specific DEX
     */
    function getAmountsOut(
        address dexRouter,
        uint256 amountIn,
        address tokenIn,
        address tokenOut
    ) external view returns (uint256 amountOut, address[] memory path) {
        path = new address[](2);
        path[0] = tokenIn;
        path[1] = tokenOut;

        // Call DEX router's getAmountsOut function
        IUniswapV2Router router = IUniswapV2Router(dexRouter);
        uint256[] memory amounts = router.getAmountsOut(amountIn, path);

        return (amounts[amounts.length - 1], path);
    }

    /**
     * @dev Swap tokens using the best available price
     */
    function swapWithBestPrice(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 minAmountOut,
        uint256 deadline
    ) external nonReentrant returns (uint256 amountOut) {
        require(amountIn > 0, "Amount must be > 0");
        require(deadline >= block.timestamp, "Deadline expired");

        // Get best quote
        Quote memory quote = getBestQuote(tokenIn, tokenOut, amountIn);
        require(quote.amountOut >= minAmountOut, "Insufficient output amount");

        // Transfer tokens from user
        IERC20(tokenIn).transferFrom(msg.sender, address(this), amountIn);

        // Calculate platform fee
        uint256 feeAmount = (amountIn * platformFee) / 10000;
        uint256 swapAmount = amountIn - feeAmount;

        // Transfer fee
        if (feeAmount > 0) {
            IERC20(tokenIn).transfer(feeCollector, feeAmount);
        }

        // Approve DEX router
        IERC20(tokenIn).approve(quote.dex, swapAmount);

        // Execute swap on best DEX
        IUniswapV2Router router = IUniswapV2Router(quote.dex);
        uint256[] memory amounts = router.swapExactTokensForTokens(
            swapAmount,
            minAmountOut,
            quote.path,
            msg.sender,
            deadline
        );

        amountOut = amounts[amounts.length - 1];

        emit SwapExecuted(
            msg.sender,
            tokenIn,
            tokenOut,
            amountIn,
            amountOut,
            quote.dex
        );

        return amountOut;
    }

    /**
     * @dev Multi-hop swap through best path
     */
    function swapMultiHop(
        address[] calldata path,
        uint256 amountIn,
        uint256 minAmountOut,
        uint256 deadline
    ) external nonReentrant returns (uint256 amountOut) {
        require(path.length >= 2, "Invalid path");
        require(amountIn > 0, "Amount must be > 0");
        require(deadline >= block.timestamp, "Deadline expired");

        // Find best DEX for this path
        uint256 bestAmountOut = 0;
        address bestDEX;

        for (uint256 i = 0; i < dexRouters.length; i++) {
            try IUniswapV2Router(dexRouters[i]).getAmountsOut(amountIn, path) returns (
                uint256[] memory amounts
            ) {
                uint256 finalAmount = amounts[amounts.length - 1];
                if (finalAmount > bestAmountOut) {
                    bestAmountOut = finalAmount;
                    bestDEX = dexRouters[i];
                }
            } catch {
                continue;
            }
        }

        require(bestAmountOut >= minAmountOut, "Insufficient output amount");

        // Transfer tokens and execute swap
        IERC20(path[0]).transferFrom(msg.sender, address(this), amountIn);

        uint256 feeAmount = (amountIn * platformFee) / 10000;
        uint256 swapAmount = amountIn - feeAmount;

        if (feeAmount > 0) {
            IERC20(path[0]).transfer(feeCollector, feeAmount);
        }

        IERC20(path[0]).approve(bestDEX, swapAmount);

        IUniswapV2Router router = IUniswapV2Router(bestDEX);
        uint256[] memory amounts = router.swapExactTokensForTokens(
            swapAmount,
            minAmountOut,
            path,
            msg.sender,
            deadline
        );

        return amounts[amounts.length - 1];
    }

    /**
     * @dev Update platform fee
     */
    function updatePlatformFee(uint256 _newFee) external onlyOwner {
        require(_newFee <= 100, "Fee too high"); // Max 1%
        platformFee = _newFee;
        emit FeeUpdated(_newFee);
    }

    /**
     * @dev Update fee collector address
     */
    function updateFeeCollector(address _newCollector) external onlyOwner {
        require(_newCollector != address(0), "Invalid address");
        feeCollector = _newCollector;
    }

    /**
     * @dev Get all supported DEX routers
     */
    function getSupportedDEXes() external view returns (address[] memory) {
        return dexRouters;
    }

    /**
     * @dev Emergency token recovery
     */
    function recoverTokens(address token, uint256 amount) external onlyOwner {
        IERC20(token).transfer(owner(), amount);
    }
}

/**
 * @dev Uniswap V2 Router interface
 */
interface IUniswapV2Router {
    function getAmountsOut(uint256 amountIn, address[] memory path)
        external
        view
        returns (uint256[] memory amounts);

    function swapExactTokensForTokens(
        uint256 amountIn,
        uint256 amountOutMin,
        address[] calldata path,
        address to,
        uint256 deadline
    ) external returns (uint256[] memory amounts);
}

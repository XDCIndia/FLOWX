// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title FlowXPool — minimal constant-product AMM for TXDC <-> tUSDC on
///         XDC Apothem testnet. No LP token: liquidity is operator-managed
///         by a single `lp` address (the FlowX treasury). Swap fee is a
///         fixed 0.3% (997/1000), taken by reducing the effective input.
///
/// Integer conventions:
///   reserve0 = TXDC balance in wei (18 decimals, native)
///   reserve1 = ERC-20 token balance in raw token units (per token.decimals())
///
/// Quote math (mirrored 1:1 in internal/routing/swap_amm.go — keep in sync):
///   amountInWithFee = amountIn * 997
///   amountOut = amountInWithFee * reserveOut / (reserveIn * 1000 + amountInWithFee)
contract FlowXPool {
    /// @notice ERC-20 token traded against TXDC (tUSDC).
    address public immutable token;
    /// @notice Liquidity operator allowed to add/remove liquidity.
    address public immutable lp;

    uint256 public reserve0; // TXDC (wei)
    uint256 public reserve1; // token (raw)
    uint256 public kLast;    // reserve0 * reserve1 after last state change

    uint256 public constant FEE_NUM = 997; // 0.3% fee
    uint256 public constant FEE_DEN = 1000;

    event LiquidityAdded(address indexed operator, uint256 txdcAmount, uint256 tokenAmount);
    event LiquidityRemoved(address indexed operator, uint256 txdcAmount, uint256 tokenAmount);
    event Swap(
        address indexed sender,
        address indexed to,
        uint256 txdcIn,
        uint256 txdcOut,
        uint256 tokenIn,
        uint256 tokenOut
    );

    error ZeroAmount();
    error InsufficientLiquidity();
    error SlippageExceeded();
    error Expired();
    error TransferFailed();
    error NotLP();

    modifier onlyLP() {
        if (msg.sender != lp) revert NotLP();
        _;
    }

    constructor(address _token, address _lp) {
        require(_token != address(0) && _lp != address(0), "zero address");
        token = _token;
        lp = _lp;
    }

    /// @notice Accept stray TXDC (counted on next addLiquidity/swap).
    receive() external payable {}

    // ------------------------------------------------------------------
    // Liquidity
    // ------------------------------------------------------------------

    /// @notice Add liquidity: msg.value TXDC (wei) plus `tokenAmount` raw
    ///         tokens pulled from the caller (lp) via transferFrom. Caller
    ///         must have approved this contract for tokenAmount first.
    function addLiquidity(uint256 tokenAmount) external payable onlyLP {
        if (msg.value == 0 || tokenAmount == 0) revert ZeroAmount();
        _safeTransferFrom(token, msg.sender, address(this), tokenAmount);
        reserve0 += msg.value;
        reserve1 += tokenAmount;
        kLast = reserve0 * reserve1;
        emit LiquidityAdded(msg.sender, msg.value, tokenAmount);
    }

    /// @notice Remove `liquidityBps` basis points (of 10000) of both
    ///         reserves, sending them back to the lp.
    function removeLiquidity(uint256 liquidityBps) external onlyLP {
        if (liquidityBps == 0 || liquidityBps > 10000) revert ZeroAmount();
        uint256 txdcOut = (reserve0 * liquidityBps) / 10000;
        uint256 tokenOut = (reserve1 * liquidityBps) / 10000;
        reserve0 -= txdcOut;
        reserve1 -= tokenOut;
        kLast = reserve0 * reserve1;
        _safeTransfer(token, msg.sender, tokenOut);
        (bool ok, ) = msg.sender.call{value: txdcOut}("");
        if (!ok) revert TransferFailed();
        emit LiquidityRemoved(msg.sender, txdcOut, tokenOut);
    }

    // ------------------------------------------------------------------
    // Swaps
    // ------------------------------------------------------------------

    /// @notice Swap an exact TXDC input for tokens, sent to `to`.
    /// @param minTokensOut Minimum tokens out (raw units), after 0.3% fee.
    function swapExactTXDCForTokens(
        uint256 minTokensOut,
        address to,
        uint256 deadline
    ) external payable returns (uint256 tokensOut) {
        if (block.timestamp > deadline) revert Expired();
        if (msg.value == 0) revert ZeroAmount();
        tokensOut = getAmountOut(msg.value, reserve0, reserve1);
        if (tokensOut > reserve1) revert InsufficientLiquidity();
        if (tokensOut < minTokensOut) revert SlippageExceeded();
        reserve0 += msg.value;
        reserve1 -= tokensOut;
        kLast = reserve0 * reserve1;
        _safeTransfer(token, to, tokensOut);
        emit Swap(msg.sender, to, msg.value, 0, 0, tokensOut);
    }

    /// @notice Swap an exact token input for TXDC, sent to `to`.
    /// @param tokenAmount  Token input in raw units, pulled from caller.
    /// @param minTXDCOut   Minimum TXDC out (wei), after 0.3% fee.
    function swapExactTokensForTXDC(
        uint256 tokenAmount,
        uint256 minTXDCOut,
        address to,
        uint256 deadline
    ) external returns (uint256 txdcOut) {
        if (block.timestamp > deadline) revert Expired();
        if (tokenAmount == 0) revert ZeroAmount();
        _safeTransferFrom(token, msg.sender, address(this), tokenAmount);
        txdcOut = getAmountOut(tokenAmount, reserve1, reserve0);
        if (txdcOut > reserve0) revert InsufficientLiquidity();
        if (txdcOut < minTXDCOut) revert SlippageExceeded();
        reserve1 += tokenAmount;
        reserve0 -= txdcOut;
        kLast = reserve0 * reserve1;
        (bool ok, ) = to.call{value: txdcOut}("");
        if (!ok) revert TransferFailed();
        emit Swap(msg.sender, to, 0, txdcOut, tokenAmount, 0);
    }

    // ------------------------------------------------------------------
    // Views
    // ------------------------------------------------------------------

    /// @notice Current reserves: (TXDC wei, token raw units).
    function reserves() external view returns (uint256, uint256) {
        return (reserve0, reserve1);
    }

    /// @notice Constant-product quote with the 0.3% fee. Pure function —
    ///         off-chain quoters (Go route) must mirror this exactly.
    function getAmountOut(
        uint256 amountIn,
        uint256 reserveIn,
        uint256 reserveOut
    ) public pure returns (uint256) {
        if (amountIn == 0 || reserveIn == 0 || reserveOut == 0) revert ZeroAmount();
        uint256 amountInWithFee = amountIn * FEE_NUM;
        uint256 numerator = amountInWithFee * reserveOut;
        uint256 denominator = reserveIn * FEE_DEN + amountInWithFee;
        return numerator / denominator;
    }

    // ------------------------------------------------------------------
    // Internals
    // ------------------------------------------------------------------

    function _safeTransfer(address tok, address to, uint256 amount) internal {
        (bool ok, bytes memory data) = tok.call(
            abi.encodeWithSignature("transfer(address,uint256)", to, amount)
        );
        if (!ok || (data.length > 0 && !abi.decode(data, (bool)))) revert TransferFailed();
    }

    function _safeTransferFrom(address tok, address from, address to, uint256 amount) internal {
        (bool ok, bytes memory data) = tok.call(
            abi.encodeWithSignature("transferFrom(address,address,uint256)", from, to, amount)
        );
        if (!ok || (data.length > 0 && !abi.decode(data, (bool)))) revert TransferFailed();
    }
}

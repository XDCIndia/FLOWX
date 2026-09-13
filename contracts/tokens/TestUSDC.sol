// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title FlowX Test USD (tUSDC)
/// @notice Minimal ERC-20 testnet stablecoin for XDC Apothem, 6 decimals.
/// OpenZeppelin-style surface: name/symbol/decimals/totalSupply,
/// balanceOf/transfer/allowance/approve/transferFrom, Transfer + Approval
/// events. Initial supply is minted to the deployer. No mint/burn/pause
/// beyond that — keep it boring.
contract TestUSDC {
    string public constant name = "FlowX Test USD";
    string public constant symbol = "tUSDC";
    uint8 public constant decimals = 6;

    uint256 public totalSupply;

    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    constructor(uint256 initialSupply) {
        totalSupply = initialSupply;
        balanceOf[msg.sender] = initialSupply;
        emit Transfer(address(0), msg.sender, initialSupply);
    }

    function transfer(address to, uint256 value) external returns (bool) {
        require(to != address(0), "ERC20: transfer to zero address");
        uint256 fromBalance = balanceOf[msg.sender];
        require(fromBalance >= value, "ERC20: insufficient balance");
        unchecked {
            balanceOf[msg.sender] = fromBalance - value;
            balanceOf[to] += value;
        }
        emit Transfer(msg.sender, to, value);
        return true;
    }

    function approve(address spender, uint256 value) external returns (bool) {
        require(spender != address(0), "ERC20: approve to zero address");
        allowance[msg.sender][spender] = value;
        emit Approval(msg.sender, spender, value);
        return true;
    }

    function transferFrom(address from, address to, uint256 value) external returns (bool) {
        require(to != address(0), "ERC20: transfer to zero address");
        uint256 allowed = allowance[from][msg.sender];
        if (allowed != type(uint256).max) {
            require(allowed >= value, "ERC20: insufficient allowance");
            unchecked {
                allowance[from][msg.sender] = allowed - value;
            }
        }
        uint256 fromBalance = balanceOf[from];
        require(fromBalance >= value, "ERC20: insufficient balance");
        unchecked {
            balanceOf[from] = fromBalance - value;
            balanceOf[to] += value;
        }
        emit Transfer(from, to, value);
        return true;
    }
}

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package tokens

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// TestUSDCMetaData contains all meta data concerning the TestUSDC contract.
var TestUSDCMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"initialSupply\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801562000010575f80fd5b5060405162000f7c38038062000f7c833981810160405281019062000036919062000128565b805f819055508060015f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055503373ffffffffffffffffffffffffffffffffffffffff165f73ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051620000dd919062000169565b60405180910390a35062000184565b5f80fd5b5f819050919050565b6200010481620000f0565b81146200010f575f80fd5b50565b5f815190506200012281620000f9565b92915050565b5f6020828403121562000140576200013f620000ec565b5b5f6200014f8482850162000112565b91505092915050565b6200016381620000f0565b82525050565b5f6020820190506200017e5f83018462000158565b92915050565b610dea80620001925f395ff3fe608060405234801561000f575f80fd5b5060043610610091575f3560e01c8063313ce56711610064578063313ce5671461013157806370a082311461014f57806395d89b411461017f578063a9059cbb1461019d578063dd62ed3e146101cd57610091565b806306fdde0314610095578063095ea7b3146100b357806318160ddd146100e357806323b872dd14610101575b5f80fd5b61009d6101fd565b6040516100aa91906109dd565b60405180910390f35b6100cd60048036038101906100c89190610a8e565b610236565b6040516100da9190610ae6565b60405180910390f35b6100eb610391565b6040516100f89190610b0e565b60405180910390f35b61011b60048036038101906101169190610b27565b610396565b6040516101289190610ae6565b60405180910390f35b6101396106ef565b6040516101469190610b92565b60405180910390f35b61016960048036038101906101649190610bab565b6106f4565b6040516101769190610b0e565b60405180910390f35b610187610709565b60405161019491906109dd565b60405180910390f35b6101b760048036038101906101b29190610a8e565b610742565b6040516101c49190610ae6565b60405180910390f35b6101e760048036038101906101e29190610bd6565b610933565b6040516101f49190610b0e565b60405180910390f35b6040518060400160405280600e81526020017f466c6f775820546573742055534400000000000000000000000000000000000081525081565b5f8073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036102a5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161029c90610c5e565b60405180910390fd5b8160025f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258460405161037f9190610b0e565b60405180910390a36001905092915050565b5f5481565b5f8073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610405576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103fc90610cc6565b60405180910390fd5b5f60025f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205490507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811461056b57828110156104eb576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104e290610d2e565b60405180910390fd5b82810360025f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055505b5f60015f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050838110156105ef576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105e690610d96565b60405180910390fd5b83810360015f8873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508360015f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055508473ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef866040516106da9190610b0e565b60405180910390a36001925050509392505050565b600681565b6001602052805f5260405f205f915090505481565b6040518060400160405280600581526020017f745553444300000000000000000000000000000000000000000000000000000081525081565b5f8073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036107b1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107a890610cc6565b60405180910390fd5b5f60015f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905082811015610835576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161082c90610d96565b60405180910390fd5b82810360015f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508260015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055508373ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef856040516109209190610b0e565b60405180910390a3600191505092915050565b6002602052815f5260405f20602052805f5260405f205f91509150505481565b5f81519050919050565b5f82825260208201905092915050565b5f5b8381101561098a57808201518184015260208101905061096f565b5f8484015250505050565b5f601f19601f8301169050919050565b5f6109af82610953565b6109b9818561095d565b93506109c981856020860161096d565b6109d281610995565b840191505092915050565b5f6020820190508181035f8301526109f581846109a5565b905092915050565b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610a2a82610a01565b9050919050565b610a3a81610a20565b8114610a44575f80fd5b50565b5f81359050610a5581610a31565b92915050565b5f819050919050565b610a6d81610a5b565b8114610a77575f80fd5b50565b5f81359050610a8881610a64565b92915050565b5f8060408385031215610aa457610aa36109fd565b5b5f610ab185828601610a47565b9250506020610ac285828601610a7a565b9150509250929050565b5f8115159050919050565b610ae081610acc565b82525050565b5f602082019050610af95f830184610ad7565b92915050565b610b0881610a5b565b82525050565b5f602082019050610b215f830184610aff565b92915050565b5f805f60608486031215610b3e57610b3d6109fd565b5b5f610b4b86828701610a47565b9350506020610b5c86828701610a47565b9250506040610b6d86828701610a7a565b9150509250925092565b5f60ff82169050919050565b610b8c81610b77565b82525050565b5f602082019050610ba55f830184610b83565b92915050565b5f60208284031215610bc057610bbf6109fd565b5b5f610bcd84828501610a47565b91505092915050565b5f8060408385031215610bec57610beb6109fd565b5b5f610bf985828601610a47565b9250506020610c0a85828601610a47565b9150509250929050565b7f45524332303a20617070726f766520746f207a65726f206164647265737300005f82015250565b5f610c48601e8361095d565b9150610c5382610c14565b602082019050919050565b5f6020820190508181035f830152610c7581610c3c565b9050919050565b7f45524332303a207472616e7366657220746f207a65726f2061646472657373005f82015250565b5f610cb0601f8361095d565b9150610cbb82610c7c565b602082019050919050565b5f6020820190508181035f830152610cdd81610ca4565b9050919050565b7f45524332303a20696e73756666696369656e7420616c6c6f77616e63650000005f82015250565b5f610d18601d8361095d565b9150610d2382610ce4565b602082019050919050565b5f6020820190508181035f830152610d4581610d0c565b9050919050565b7f45524332303a20696e73756666696369656e742062616c616e636500000000005f82015250565b5f610d80601b8361095d565b9150610d8b82610d4c565b602082019050919050565b5f6020820190508181035f830152610dad81610d74565b905091905056fea264697066735822122062cb4225a5671cd7bff4d7750cb608344cb0a66c49084460316d252ba5a4d8cc64736f6c63430008180033",
}

// TestUSDCABI is the input ABI used to generate the binding from.
// Deprecated: Use TestUSDCMetaData.ABI instead.
var TestUSDCABI = TestUSDCMetaData.ABI

// TestUSDCBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use TestUSDCMetaData.Bin instead.
var TestUSDCBin = TestUSDCMetaData.Bin

// DeployTestUSDC deploys a new Ethereum contract, binding an instance of TestUSDC to it.
func DeployTestUSDC(auth *bind.TransactOpts, backend bind.ContractBackend, initialSupply *big.Int) (common.Address, *types.Transaction, *TestUSDC, error) {
	parsed, err := TestUSDCMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TestUSDCBin), backend, initialSupply)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &TestUSDC{TestUSDCCaller: TestUSDCCaller{contract: contract}, TestUSDCTransactor: TestUSDCTransactor{contract: contract}, TestUSDCFilterer: TestUSDCFilterer{contract: contract}}, nil
}

// TestUSDC is an auto generated Go binding around an Ethereum contract.
type TestUSDC struct {
	TestUSDCCaller     // Read-only binding to the contract
	TestUSDCTransactor // Write-only binding to the contract
	TestUSDCFilterer   // Log filterer for contract events
}

// TestUSDCCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestUSDCCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestUSDCTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestUSDCTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestUSDCFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestUSDCFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestUSDCSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestUSDCSession struct {
	Contract     *TestUSDC         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestUSDCCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestUSDCCallerSession struct {
	Contract *TestUSDCCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// TestUSDCTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestUSDCTransactorSession struct {
	Contract     *TestUSDCTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// TestUSDCRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestUSDCRaw struct {
	Contract *TestUSDC // Generic contract binding to access the raw methods on
}

// TestUSDCCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestUSDCCallerRaw struct {
	Contract *TestUSDCCaller // Generic read-only contract binding to access the raw methods on
}

// TestUSDCTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestUSDCTransactorRaw struct {
	Contract *TestUSDCTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestUSDC creates a new instance of TestUSDC, bound to a specific deployed contract.
func NewTestUSDC(address common.Address, backend bind.ContractBackend) (*TestUSDC, error) {
	contract, err := bindTestUSDC(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestUSDC{TestUSDCCaller: TestUSDCCaller{contract: contract}, TestUSDCTransactor: TestUSDCTransactor{contract: contract}, TestUSDCFilterer: TestUSDCFilterer{contract: contract}}, nil
}

// NewTestUSDCCaller creates a new read-only instance of TestUSDC, bound to a specific deployed contract.
func NewTestUSDCCaller(address common.Address, caller bind.ContractCaller) (*TestUSDCCaller, error) {
	contract, err := bindTestUSDC(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestUSDCCaller{contract: contract}, nil
}

// NewTestUSDCTransactor creates a new write-only instance of TestUSDC, bound to a specific deployed contract.
func NewTestUSDCTransactor(address common.Address, transactor bind.ContractTransactor) (*TestUSDCTransactor, error) {
	contract, err := bindTestUSDC(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestUSDCTransactor{contract: contract}, nil
}

// NewTestUSDCFilterer creates a new log filterer instance of TestUSDC, bound to a specific deployed contract.
func NewTestUSDCFilterer(address common.Address, filterer bind.ContractFilterer) (*TestUSDCFilterer, error) {
	contract, err := bindTestUSDC(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestUSDCFilterer{contract: contract}, nil
}

// bindTestUSDC binds a generic wrapper to an already deployed contract.
func bindTestUSDC(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TestUSDCMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestUSDC *TestUSDCRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestUSDC.Contract.TestUSDCCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestUSDC *TestUSDCRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestUSDC.Contract.TestUSDCTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestUSDC *TestUSDCRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestUSDC.Contract.TestUSDCTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestUSDC *TestUSDCCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestUSDC.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestUSDC *TestUSDCTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestUSDC.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestUSDC *TestUSDCTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestUSDC.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_TestUSDC *TestUSDCCaller) Allowance(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _TestUSDC.contract.Call(opts, &out, "allowance", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_TestUSDC *TestUSDCSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _TestUSDC.Contract.Allowance(&_TestUSDC.CallOpts, arg0, arg1)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_TestUSDC *TestUSDCCallerSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _TestUSDC.Contract.Allowance(&_TestUSDC.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_TestUSDC *TestUSDCCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _TestUSDC.contract.Call(opts, &out, "balanceOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_TestUSDC *TestUSDCSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _TestUSDC.Contract.BalanceOf(&_TestUSDC.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_TestUSDC *TestUSDCCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _TestUSDC.Contract.BalanceOf(&_TestUSDC.CallOpts, arg0)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TestUSDC *TestUSDCCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _TestUSDC.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TestUSDC *TestUSDCSession) Decimals() (uint8, error) {
	return _TestUSDC.Contract.Decimals(&_TestUSDC.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TestUSDC *TestUSDCCallerSession) Decimals() (uint8, error) {
	return _TestUSDC.Contract.Decimals(&_TestUSDC.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TestUSDC *TestUSDCCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _TestUSDC.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TestUSDC *TestUSDCSession) Name() (string, error) {
	return _TestUSDC.Contract.Name(&_TestUSDC.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TestUSDC *TestUSDCCallerSession) Name() (string, error) {
	return _TestUSDC.Contract.Name(&_TestUSDC.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TestUSDC *TestUSDCCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _TestUSDC.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TestUSDC *TestUSDCSession) Symbol() (string, error) {
	return _TestUSDC.Contract.Symbol(&_TestUSDC.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TestUSDC *TestUSDCCallerSession) Symbol() (string, error) {
	return _TestUSDC.Contract.Symbol(&_TestUSDC.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TestUSDC *TestUSDCCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TestUSDC.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TestUSDC *TestUSDCSession) TotalSupply() (*big.Int, error) {
	return _TestUSDC.Contract.TotalSupply(&_TestUSDC.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TestUSDC *TestUSDCCallerSession) TotalSupply() (*big.Int, error) {
	return _TestUSDC.Contract.TotalSupply(&_TestUSDC.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCTransactor) Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.contract.Transact(opts, "approve", spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.Contract.Approve(&_TestUSDC.TransactOpts, spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCTransactorSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.Contract.Approve(&_TestUSDC.TransactOpts, spender, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCTransactor) Transfer(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.contract.Transact(opts, "transfer", to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.Contract.Transfer(&_TestUSDC.TransactOpts, to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCTransactorSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.Contract.Transfer(&_TestUSDC.TransactOpts, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.contract.Transact(opts, "transferFrom", from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.Contract.TransferFrom(&_TestUSDC.TransactOpts, from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_TestUSDC *TestUSDCTransactorSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _TestUSDC.Contract.TransferFrom(&_TestUSDC.TransactOpts, from, to, value)
}

// TestUSDCApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the TestUSDC contract.
type TestUSDCApprovalIterator struct {
	Event *TestUSDCApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TestUSDCApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestUSDCApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TestUSDCApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TestUSDCApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestUSDCApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestUSDCApproval represents a Approval event raised by the TestUSDC contract.
type TestUSDCApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TestUSDC *TestUSDCFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*TestUSDCApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _TestUSDC.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &TestUSDCApprovalIterator{contract: _TestUSDC.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TestUSDC *TestUSDCFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *TestUSDCApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _TestUSDC.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestUSDCApproval)
				if err := _TestUSDC.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TestUSDC *TestUSDCFilterer) ParseApproval(log types.Log) (*TestUSDCApproval, error) {
	event := new(TestUSDCApproval)
	if err := _TestUSDC.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TestUSDCTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the TestUSDC contract.
type TestUSDCTransferIterator struct {
	Event *TestUSDCTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TestUSDCTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestUSDCTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TestUSDCTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TestUSDCTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestUSDCTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestUSDCTransfer represents a Transfer event raised by the TestUSDC contract.
type TestUSDCTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TestUSDC *TestUSDCFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*TestUSDCTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _TestUSDC.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &TestUSDCTransferIterator{contract: _TestUSDC.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TestUSDC *TestUSDCFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *TestUSDCTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _TestUSDC.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestUSDCTransfer)
				if err := _TestUSDC.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TestUSDC *TestUSDCFilterer) ParseTransfer(log types.Log) (*TestUSDCTransfer, error) {
	event := new(TestUSDCTransfer)
	if err := _TestUSDC.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

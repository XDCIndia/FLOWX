// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package amm

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

// FlowXPoolMetaData contains all meta data concerning the FlowXPool contract.
var FlowXPoolMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_lp\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"Expired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientLiquidity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotLP\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SlippageExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TransferFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAmount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"txdcAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"}],\"name\":\"LiquidityAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"txdcAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"}],\"name\":\"LiquidityRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"txdcIn\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"txdcOut\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenIn\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenOut\",\"type\":\"uint256\"}],\"name\":\"Swap\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"FEE_DEN\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"FEE_NUM\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"}],\"name\":\"addLiquidity\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"reserveIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"reserveOut\",\"type\":\"uint256\"}],\"name\":\"getAmountOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"kLast\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lp\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"liquidityBps\",\"type\":\"uint256\"}],\"name\":\"removeLiquidity\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"reserve0\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"reserve1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"reserves\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minTokensOut\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"swapExactTXDCForTokens\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokensOut\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minTXDCOut\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"swapExactTokensForTXDC\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"txdcOut\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60c060405234801561000f575f80fd5b50604051610dcd380380610dcd83398101604081905261002e916100bf565b6001600160a01b0382161580159061004e57506001600160a01b03811615155b61008d5760405162461bcd60e51b815260206004820152600c60248201526b7a65726f206164647265737360a01b604482015260640160405180910390fd5b6001600160a01b039182166080521660a0526100f0565b80516001600160a01b03811681146100ba575f80fd5b919050565b5f80604083850312156100d0575f80fd5b6100d9836100a4565b91506100e7602084016100a4565b90509250929050565b60805160a051610c9261013b5f395f818161013801528181610454015261073a01525f818161023a015281816103c9015281816104c0015281816105ab01526108230152610c925ff3fe6080604052600436106100c2575f3560e01c806353cd0edd1161007c57806375172a8b1161005757806375172a8b146101e45780639c8f9f231461020a578063fc0c546a14610229578063fe08b58c1461025c575f80fd5b806353cd0edd1461019b5780635a76f25e146101ba5780637464fc3d146101cf575f80fd5b8063054d50d4146100cd5780631558a352146100ff5780631605b17f14610112578063313c06a014610127578063443cb4bc1461017257806351c6590a14610186575f80fd5b366100c957005b5f80fd5b3480156100d8575f80fd5b506100ec6100e7366004610acd565b610271565b6040519081526020015b60405180910390f35b6100ec61010d366004610b11565b6102ef565b34801561011d575f80fd5b506100ec6103e881565b348015610132575f80fd5b5061015a7f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020016100f6565b34801561017d575f80fd5b506100ec5f5481565b610199610194366004610b43565b610449565b005b3480156101a6575f80fd5b506100ec6101b5366004610b5a565b610564565b3480156101c5575f80fd5b506100ec60015481565b3480156101da575f80fd5b506100ec60025481565b3480156101ef575f80fd5b505f54600154604080519283526020830191909152016100f6565b348015610215575f80fd5b50610199610224366004610b43565b61072f565b348015610234575f80fd5b5061015a7f000000000000000000000000000000000000000000000000000000000000000081565b348015610267575f80fd5b506100ec6103e581565b5f83158061027d575082155b80610286575081155b156102a457604051631f2a200560e01b815260040160405180910390fd5b5f6102b16103e586610ba8565b90505f6102be8483610ba8565b90505f826102ce6103e888610ba8565b6102d89190610bc5565b90506102e48183610bd8565b979650505050505050565b5f8142111561031157604051630407b05b60e31b815260040160405180910390fd5b345f0361033157604051631f2a200560e01b815260040160405180910390fd5b61033f345f54600154610271565b90506001548111156103645760405163bb55fd2760e01b815260040160405180910390fd5b8381101561038557604051638199f5f360e01b815260040160405180910390fd5b345f808282546103959190610bc5565b925050819055508060015f8282546103ad9190610bf7565b90915550506001545f546103c19190610ba8565b6002556103ef7f000000000000000000000000000000000000000000000000000000000000000084836108f0565b604080513481525f60208201819052818301526060810183905290516001600160a01b0385169133917fb3e2773606abfd36b5bd91394b3a54d1398336c65005baf7bf7a05efeffaf75b9181900360800190a39392505050565b336001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016146104925760405163021ff9c960e51b815260040160405180910390fd5b34158061049d575080155b156104bb57604051631f2a200560e01b815260040160405180910390fd5b6104e77f00000000000000000000000000000000000000000000000000000000000000003330846109da565b345f808282546104f79190610bc5565b925050819055508060015f82825461050f9190610bc5565b90915550506001545f546105239190610ba8565b600255604080513481526020810183905233917fac1d76749e5447b7b16f5ab61447e1bd502f3bb4807af3b28e620d1700a6ee45910160405180910390a250565b5f8142111561058657604051630407b05b60e31b815260040160405180910390fd5b845f036105a657604051631f2a200560e01b815260040160405180910390fd5b6105d27f00000000000000000000000000000000000000000000000000000000000000003330886109da565b6105e0856001545f54610271565b90505f548111156106045760405163bb55fd2760e01b815260040160405180910390fd5b8381101561062557604051638199f5f360e01b815260040160405180910390fd5b8460015f8282546106369190610bc5565b92505081905550805f8082825461064d9190610bf7565b90915550506001545f546106619190610ba8565b6002556040515f906001600160a01b0385169083908381818185875af1925050503d805f81146106ac576040519150601f19603f3d011682016040523d82523d5f602084013e6106b1565b606091505b50509050806106d3576040516312171d8360e31b815260040160405180910390fd5b604080515f80825260208201859052818301899052606082015290516001600160a01b0386169133917fb3e2773606abfd36b5bd91394b3a54d1398336c65005baf7bf7a05efeffaf75b9181900360800190a350949350505050565b336001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016146107785760405163021ff9c960e51b815260040160405180910390fd5b801580610786575061271081115b156107a457604051631f2a200560e01b815260040160405180910390fd5b5f612710825f546107b59190610ba8565b6107bf9190610bd8565b90505f612710836001546107d39190610ba8565b6107dd9190610bd8565b9050815f808282546107ef9190610bf7565b925050819055508060015f8282546108079190610bf7565b90915550506001545f5461081b9190610ba8565b6002556108497f000000000000000000000000000000000000000000000000000000000000000033836108f0565b6040515f90339084908381818185875af1925050503d805f8114610888576040519150601f19603f3d011682016040523d82523d5f602084013e61088d565b606091505b50509050806108af576040516312171d8360e31b815260040160405180910390fd5b604080518481526020810184905233917f96cd817c6329656790ef8fba7675405193677d39619571282f5e21f3a98cd059910160405180910390a250505050565b6040516001600160a01b038381166024830152604482018390525f91829186169060640160408051601f198184030181529181526020820180516001600160e01b031663a9059cbb60e01b179052516109499190610c0a565b5f604051808303815f865af19150503d805f8114610982576040519150601f19603f3d011682016040523d82523d5f602084013e610987565b606091505b50915091508115806109b557505f81511180156109b55750808060200190518101906109b39190610c36565b155b156109d3576040516312171d8360e31b815260040160405180910390fd5b5050505050565b6040516001600160a01b0384811660248301528381166044830152606482018390525f91829187169060840160408051601f198184030181529181526020820180516001600160e01b03166323b872dd60e01b17905251610a3b9190610c0a565b5f604051808303815f865af19150503d805f8114610a74576040519150601f19603f3d011682016040523d82523d5f602084013e610a79565b606091505b5091509150811580610aa757505f8151118015610aa7575080806020019051810190610aa59190610c36565b155b15610ac5576040516312171d8360e31b815260040160405180910390fd5b505050505050565b5f805f60608486031215610adf575f80fd5b505081359360208301359350604090920135919050565b80356001600160a01b0381168114610b0c575f80fd5b919050565b5f805f60608486031215610b23575f80fd5b83359250610b3360208501610af6565b9150604084013590509250925092565b5f60208284031215610b53575f80fd5b5035919050565b5f805f8060808587031215610b6d575f80fd5b8435935060208501359250610b8460408601610af6565b9396929550929360600135925050565b634e487b7160e01b5f52601160045260245ffd5b8082028115828204841417610bbf57610bbf610b94565b92915050565b80820180821115610bbf57610bbf610b94565b5f82610bf257634e487b7160e01b5f52601260045260245ffd5b500490565b81810381811115610bbf57610bbf610b94565b5f82515f5b81811015610c295760208186018101518583015201610c0f565b505f920191825250919050565b5f60208284031215610c46575f80fd5b81518015158114610c55575f80fd5b939250505056fea2646970667358221220f6e889fe4841a02e3769ab6f9c987029886a7bbdee15321879b529db5913d3a464736f6c63430008180033",
}

// FlowXPoolABI is the input ABI used to generate the binding from.
// Deprecated: Use FlowXPoolMetaData.ABI instead.
var FlowXPoolABI = FlowXPoolMetaData.ABI

// FlowXPoolBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use FlowXPoolMetaData.Bin instead.
var FlowXPoolBin = FlowXPoolMetaData.Bin

// DeployFlowXPool deploys a new Ethereum contract, binding an instance of FlowXPool to it.
func DeployFlowXPool(auth *bind.TransactOpts, backend bind.ContractBackend, _token common.Address, _lp common.Address) (common.Address, *types.Transaction, *FlowXPool, error) {
	parsed, err := FlowXPoolMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(FlowXPoolBin), backend, _token, _lp)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &FlowXPool{FlowXPoolCaller: FlowXPoolCaller{contract: contract}, FlowXPoolTransactor: FlowXPoolTransactor{contract: contract}, FlowXPoolFilterer: FlowXPoolFilterer{contract: contract}}, nil
}

// FlowXPool is an auto generated Go binding around an Ethereum contract.
type FlowXPool struct {
	FlowXPoolCaller     // Read-only binding to the contract
	FlowXPoolTransactor // Write-only binding to the contract
	FlowXPoolFilterer   // Log filterer for contract events
}

// FlowXPoolCaller is an auto generated read-only Go binding around an Ethereum contract.
type FlowXPoolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlowXPoolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FlowXPoolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlowXPoolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FlowXPoolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlowXPoolSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FlowXPoolSession struct {
	Contract     *FlowXPool        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FlowXPoolCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FlowXPoolCallerSession struct {
	Contract *FlowXPoolCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// FlowXPoolTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FlowXPoolTransactorSession struct {
	Contract     *FlowXPoolTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// FlowXPoolRaw is an auto generated low-level Go binding around an Ethereum contract.
type FlowXPoolRaw struct {
	Contract *FlowXPool // Generic contract binding to access the raw methods on
}

// FlowXPoolCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FlowXPoolCallerRaw struct {
	Contract *FlowXPoolCaller // Generic read-only contract binding to access the raw methods on
}

// FlowXPoolTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FlowXPoolTransactorRaw struct {
	Contract *FlowXPoolTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFlowXPool creates a new instance of FlowXPool, bound to a specific deployed contract.
func NewFlowXPool(address common.Address, backend bind.ContractBackend) (*FlowXPool, error) {
	contract, err := bindFlowXPool(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FlowXPool{FlowXPoolCaller: FlowXPoolCaller{contract: contract}, FlowXPoolTransactor: FlowXPoolTransactor{contract: contract}, FlowXPoolFilterer: FlowXPoolFilterer{contract: contract}}, nil
}

// NewFlowXPoolCaller creates a new read-only instance of FlowXPool, bound to a specific deployed contract.
func NewFlowXPoolCaller(address common.Address, caller bind.ContractCaller) (*FlowXPoolCaller, error) {
	contract, err := bindFlowXPool(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FlowXPoolCaller{contract: contract}, nil
}

// NewFlowXPoolTransactor creates a new write-only instance of FlowXPool, bound to a specific deployed contract.
func NewFlowXPoolTransactor(address common.Address, transactor bind.ContractTransactor) (*FlowXPoolTransactor, error) {
	contract, err := bindFlowXPool(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FlowXPoolTransactor{contract: contract}, nil
}

// NewFlowXPoolFilterer creates a new log filterer instance of FlowXPool, bound to a specific deployed contract.
func NewFlowXPoolFilterer(address common.Address, filterer bind.ContractFilterer) (*FlowXPoolFilterer, error) {
	contract, err := bindFlowXPool(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FlowXPoolFilterer{contract: contract}, nil
}

// bindFlowXPool binds a generic wrapper to an already deployed contract.
func bindFlowXPool(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := FlowXPoolMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FlowXPool *FlowXPoolRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FlowXPool.Contract.FlowXPoolCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FlowXPool *FlowXPoolRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlowXPool.Contract.FlowXPoolTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FlowXPool *FlowXPoolRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FlowXPool.Contract.FlowXPoolTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FlowXPool *FlowXPoolCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FlowXPool.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FlowXPool *FlowXPoolTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlowXPool.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FlowXPool *FlowXPoolTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FlowXPool.Contract.contract.Transact(opts, method, params...)
}

// FEEDEN is a free data retrieval call binding the contract method 0x1605b17f.
//
// Solidity: function FEE_DEN() view returns(uint256)
func (_FlowXPool *FlowXPoolCaller) FEEDEN(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "FEE_DEN")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FEEDEN is a free data retrieval call binding the contract method 0x1605b17f.
//
// Solidity: function FEE_DEN() view returns(uint256)
func (_FlowXPool *FlowXPoolSession) FEEDEN() (*big.Int, error) {
	return _FlowXPool.Contract.FEEDEN(&_FlowXPool.CallOpts)
}

// FEEDEN is a free data retrieval call binding the contract method 0x1605b17f.
//
// Solidity: function FEE_DEN() view returns(uint256)
func (_FlowXPool *FlowXPoolCallerSession) FEEDEN() (*big.Int, error) {
	return _FlowXPool.Contract.FEEDEN(&_FlowXPool.CallOpts)
}

// FEENUM is a free data retrieval call binding the contract method 0xfe08b58c.
//
// Solidity: function FEE_NUM() view returns(uint256)
func (_FlowXPool *FlowXPoolCaller) FEENUM(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "FEE_NUM")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FEENUM is a free data retrieval call binding the contract method 0xfe08b58c.
//
// Solidity: function FEE_NUM() view returns(uint256)
func (_FlowXPool *FlowXPoolSession) FEENUM() (*big.Int, error) {
	return _FlowXPool.Contract.FEENUM(&_FlowXPool.CallOpts)
}

// FEENUM is a free data retrieval call binding the contract method 0xfe08b58c.
//
// Solidity: function FEE_NUM() view returns(uint256)
func (_FlowXPool *FlowXPoolCallerSession) FEENUM() (*big.Int, error) {
	return _FlowXPool.Contract.FEENUM(&_FlowXPool.CallOpts)
}

// GetAmountOut is a free data retrieval call binding the contract method 0x054d50d4.
//
// Solidity: function getAmountOut(uint256 amountIn, uint256 reserveIn, uint256 reserveOut) pure returns(uint256)
func (_FlowXPool *FlowXPoolCaller) GetAmountOut(opts *bind.CallOpts, amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "getAmountOut", amountIn, reserveIn, reserveOut)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAmountOut is a free data retrieval call binding the contract method 0x054d50d4.
//
// Solidity: function getAmountOut(uint256 amountIn, uint256 reserveIn, uint256 reserveOut) pure returns(uint256)
func (_FlowXPool *FlowXPoolSession) GetAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) (*big.Int, error) {
	return _FlowXPool.Contract.GetAmountOut(&_FlowXPool.CallOpts, amountIn, reserveIn, reserveOut)
}

// GetAmountOut is a free data retrieval call binding the contract method 0x054d50d4.
//
// Solidity: function getAmountOut(uint256 amountIn, uint256 reserveIn, uint256 reserveOut) pure returns(uint256)
func (_FlowXPool *FlowXPoolCallerSession) GetAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) (*big.Int, error) {
	return _FlowXPool.Contract.GetAmountOut(&_FlowXPool.CallOpts, amountIn, reserveIn, reserveOut)
}

// KLast is a free data retrieval call binding the contract method 0x7464fc3d.
//
// Solidity: function kLast() view returns(uint256)
func (_FlowXPool *FlowXPoolCaller) KLast(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "kLast")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// KLast is a free data retrieval call binding the contract method 0x7464fc3d.
//
// Solidity: function kLast() view returns(uint256)
func (_FlowXPool *FlowXPoolSession) KLast() (*big.Int, error) {
	return _FlowXPool.Contract.KLast(&_FlowXPool.CallOpts)
}

// KLast is a free data retrieval call binding the contract method 0x7464fc3d.
//
// Solidity: function kLast() view returns(uint256)
func (_FlowXPool *FlowXPoolCallerSession) KLast() (*big.Int, error) {
	return _FlowXPool.Contract.KLast(&_FlowXPool.CallOpts)
}

// Lp is a free data retrieval call binding the contract method 0x313c06a0.
//
// Solidity: function lp() view returns(address)
func (_FlowXPool *FlowXPoolCaller) Lp(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "lp")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Lp is a free data retrieval call binding the contract method 0x313c06a0.
//
// Solidity: function lp() view returns(address)
func (_FlowXPool *FlowXPoolSession) Lp() (common.Address, error) {
	return _FlowXPool.Contract.Lp(&_FlowXPool.CallOpts)
}

// Lp is a free data retrieval call binding the contract method 0x313c06a0.
//
// Solidity: function lp() view returns(address)
func (_FlowXPool *FlowXPoolCallerSession) Lp() (common.Address, error) {
	return _FlowXPool.Contract.Lp(&_FlowXPool.CallOpts)
}

// Reserve0 is a free data retrieval call binding the contract method 0x443cb4bc.
//
// Solidity: function reserve0() view returns(uint256)
func (_FlowXPool *FlowXPoolCaller) Reserve0(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "reserve0")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Reserve0 is a free data retrieval call binding the contract method 0x443cb4bc.
//
// Solidity: function reserve0() view returns(uint256)
func (_FlowXPool *FlowXPoolSession) Reserve0() (*big.Int, error) {
	return _FlowXPool.Contract.Reserve0(&_FlowXPool.CallOpts)
}

// Reserve0 is a free data retrieval call binding the contract method 0x443cb4bc.
//
// Solidity: function reserve0() view returns(uint256)
func (_FlowXPool *FlowXPoolCallerSession) Reserve0() (*big.Int, error) {
	return _FlowXPool.Contract.Reserve0(&_FlowXPool.CallOpts)
}

// Reserve1 is a free data retrieval call binding the contract method 0x5a76f25e.
//
// Solidity: function reserve1() view returns(uint256)
func (_FlowXPool *FlowXPoolCaller) Reserve1(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "reserve1")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Reserve1 is a free data retrieval call binding the contract method 0x5a76f25e.
//
// Solidity: function reserve1() view returns(uint256)
func (_FlowXPool *FlowXPoolSession) Reserve1() (*big.Int, error) {
	return _FlowXPool.Contract.Reserve1(&_FlowXPool.CallOpts)
}

// Reserve1 is a free data retrieval call binding the contract method 0x5a76f25e.
//
// Solidity: function reserve1() view returns(uint256)
func (_FlowXPool *FlowXPoolCallerSession) Reserve1() (*big.Int, error) {
	return _FlowXPool.Contract.Reserve1(&_FlowXPool.CallOpts)
}

// Reserves is a free data retrieval call binding the contract method 0x75172a8b.
//
// Solidity: function reserves() view returns(uint256, uint256)
func (_FlowXPool *FlowXPoolCaller) Reserves(opts *bind.CallOpts) (*big.Int, *big.Int, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "reserves")

	if err != nil {
		return *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// Reserves is a free data retrieval call binding the contract method 0x75172a8b.
//
// Solidity: function reserves() view returns(uint256, uint256)
func (_FlowXPool *FlowXPoolSession) Reserves() (*big.Int, *big.Int, error) {
	return _FlowXPool.Contract.Reserves(&_FlowXPool.CallOpts)
}

// Reserves is a free data retrieval call binding the contract method 0x75172a8b.
//
// Solidity: function reserves() view returns(uint256, uint256)
func (_FlowXPool *FlowXPoolCallerSession) Reserves() (*big.Int, *big.Int, error) {
	return _FlowXPool.Contract.Reserves(&_FlowXPool.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_FlowXPool *FlowXPoolCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlowXPool.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_FlowXPool *FlowXPoolSession) Token() (common.Address, error) {
	return _FlowXPool.Contract.Token(&_FlowXPool.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_FlowXPool *FlowXPoolCallerSession) Token() (common.Address, error) {
	return _FlowXPool.Contract.Token(&_FlowXPool.CallOpts)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x51c6590a.
//
// Solidity: function addLiquidity(uint256 tokenAmount) payable returns()
func (_FlowXPool *FlowXPoolTransactor) AddLiquidity(opts *bind.TransactOpts, tokenAmount *big.Int) (*types.Transaction, error) {
	return _FlowXPool.contract.Transact(opts, "addLiquidity", tokenAmount)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x51c6590a.
//
// Solidity: function addLiquidity(uint256 tokenAmount) payable returns()
func (_FlowXPool *FlowXPoolSession) AddLiquidity(tokenAmount *big.Int) (*types.Transaction, error) {
	return _FlowXPool.Contract.AddLiquidity(&_FlowXPool.TransactOpts, tokenAmount)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x51c6590a.
//
// Solidity: function addLiquidity(uint256 tokenAmount) payable returns()
func (_FlowXPool *FlowXPoolTransactorSession) AddLiquidity(tokenAmount *big.Int) (*types.Transaction, error) {
	return _FlowXPool.Contract.AddLiquidity(&_FlowXPool.TransactOpts, tokenAmount)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x9c8f9f23.
//
// Solidity: function removeLiquidity(uint256 liquidityBps) returns()
func (_FlowXPool *FlowXPoolTransactor) RemoveLiquidity(opts *bind.TransactOpts, liquidityBps *big.Int) (*types.Transaction, error) {
	return _FlowXPool.contract.Transact(opts, "removeLiquidity", liquidityBps)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x9c8f9f23.
//
// Solidity: function removeLiquidity(uint256 liquidityBps) returns()
func (_FlowXPool *FlowXPoolSession) RemoveLiquidity(liquidityBps *big.Int) (*types.Transaction, error) {
	return _FlowXPool.Contract.RemoveLiquidity(&_FlowXPool.TransactOpts, liquidityBps)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x9c8f9f23.
//
// Solidity: function removeLiquidity(uint256 liquidityBps) returns()
func (_FlowXPool *FlowXPoolTransactorSession) RemoveLiquidity(liquidityBps *big.Int) (*types.Transaction, error) {
	return _FlowXPool.Contract.RemoveLiquidity(&_FlowXPool.TransactOpts, liquidityBps)
}

// SwapExactTXDCForTokens is a paid mutator transaction binding the contract method 0x1558a352.
//
// Solidity: function swapExactTXDCForTokens(uint256 minTokensOut, address to, uint256 deadline) payable returns(uint256 tokensOut)
func (_FlowXPool *FlowXPoolTransactor) SwapExactTXDCForTokens(opts *bind.TransactOpts, minTokensOut *big.Int, to common.Address, deadline *big.Int) (*types.Transaction, error) {
	return _FlowXPool.contract.Transact(opts, "swapExactTXDCForTokens", minTokensOut, to, deadline)
}

// SwapExactTXDCForTokens is a paid mutator transaction binding the contract method 0x1558a352.
//
// Solidity: function swapExactTXDCForTokens(uint256 minTokensOut, address to, uint256 deadline) payable returns(uint256 tokensOut)
func (_FlowXPool *FlowXPoolSession) SwapExactTXDCForTokens(minTokensOut *big.Int, to common.Address, deadline *big.Int) (*types.Transaction, error) {
	return _FlowXPool.Contract.SwapExactTXDCForTokens(&_FlowXPool.TransactOpts, minTokensOut, to, deadline)
}

// SwapExactTXDCForTokens is a paid mutator transaction binding the contract method 0x1558a352.
//
// Solidity: function swapExactTXDCForTokens(uint256 minTokensOut, address to, uint256 deadline) payable returns(uint256 tokensOut)
func (_FlowXPool *FlowXPoolTransactorSession) SwapExactTXDCForTokens(minTokensOut *big.Int, to common.Address, deadline *big.Int) (*types.Transaction, error) {
	return _FlowXPool.Contract.SwapExactTXDCForTokens(&_FlowXPool.TransactOpts, minTokensOut, to, deadline)
}

// SwapExactTokensForTXDC is a paid mutator transaction binding the contract method 0x53cd0edd.
//
// Solidity: function swapExactTokensForTXDC(uint256 tokenAmount, uint256 minTXDCOut, address to, uint256 deadline) returns(uint256 txdcOut)
func (_FlowXPool *FlowXPoolTransactor) SwapExactTokensForTXDC(opts *bind.TransactOpts, tokenAmount *big.Int, minTXDCOut *big.Int, to common.Address, deadline *big.Int) (*types.Transaction, error) {
	return _FlowXPool.contract.Transact(opts, "swapExactTokensForTXDC", tokenAmount, minTXDCOut, to, deadline)
}

// SwapExactTokensForTXDC is a paid mutator transaction binding the contract method 0x53cd0edd.
//
// Solidity: function swapExactTokensForTXDC(uint256 tokenAmount, uint256 minTXDCOut, address to, uint256 deadline) returns(uint256 txdcOut)
func (_FlowXPool *FlowXPoolSession) SwapExactTokensForTXDC(tokenAmount *big.Int, minTXDCOut *big.Int, to common.Address, deadline *big.Int) (*types.Transaction, error) {
	return _FlowXPool.Contract.SwapExactTokensForTXDC(&_FlowXPool.TransactOpts, tokenAmount, minTXDCOut, to, deadline)
}

// SwapExactTokensForTXDC is a paid mutator transaction binding the contract method 0x53cd0edd.
//
// Solidity: function swapExactTokensForTXDC(uint256 tokenAmount, uint256 minTXDCOut, address to, uint256 deadline) returns(uint256 txdcOut)
func (_FlowXPool *FlowXPoolTransactorSession) SwapExactTokensForTXDC(tokenAmount *big.Int, minTXDCOut *big.Int, to common.Address, deadline *big.Int) (*types.Transaction, error) {
	return _FlowXPool.Contract.SwapExactTokensForTXDC(&_FlowXPool.TransactOpts, tokenAmount, minTXDCOut, to, deadline)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_FlowXPool *FlowXPoolTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlowXPool.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_FlowXPool *FlowXPoolSession) Receive() (*types.Transaction, error) {
	return _FlowXPool.Contract.Receive(&_FlowXPool.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_FlowXPool *FlowXPoolTransactorSession) Receive() (*types.Transaction, error) {
	return _FlowXPool.Contract.Receive(&_FlowXPool.TransactOpts)
}

// FlowXPoolLiquidityAddedIterator is returned from FilterLiquidityAdded and is used to iterate over the raw logs and unpacked data for LiquidityAdded events raised by the FlowXPool contract.
type FlowXPoolLiquidityAddedIterator struct {
	Event *FlowXPoolLiquidityAdded // Event containing the contract specifics and raw log

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
func (it *FlowXPoolLiquidityAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlowXPoolLiquidityAdded)
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
		it.Event = new(FlowXPoolLiquidityAdded)
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
func (it *FlowXPoolLiquidityAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlowXPoolLiquidityAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlowXPoolLiquidityAdded represents a LiquidityAdded event raised by the FlowXPool contract.
type FlowXPoolLiquidityAdded struct {
	Operator    common.Address
	TxdcAmount  *big.Int
	TokenAmount *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterLiquidityAdded is a free log retrieval operation binding the contract event 0xac1d76749e5447b7b16f5ab61447e1bd502f3bb4807af3b28e620d1700a6ee45.
//
// Solidity: event LiquidityAdded(address indexed operator, uint256 txdcAmount, uint256 tokenAmount)
func (_FlowXPool *FlowXPoolFilterer) FilterLiquidityAdded(opts *bind.FilterOpts, operator []common.Address) (*FlowXPoolLiquidityAddedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _FlowXPool.contract.FilterLogs(opts, "LiquidityAdded", operatorRule)
	if err != nil {
		return nil, err
	}
	return &FlowXPoolLiquidityAddedIterator{contract: _FlowXPool.contract, event: "LiquidityAdded", logs: logs, sub: sub}, nil
}

// WatchLiquidityAdded is a free log subscription operation binding the contract event 0xac1d76749e5447b7b16f5ab61447e1bd502f3bb4807af3b28e620d1700a6ee45.
//
// Solidity: event LiquidityAdded(address indexed operator, uint256 txdcAmount, uint256 tokenAmount)
func (_FlowXPool *FlowXPoolFilterer) WatchLiquidityAdded(opts *bind.WatchOpts, sink chan<- *FlowXPoolLiquidityAdded, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _FlowXPool.contract.WatchLogs(opts, "LiquidityAdded", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlowXPoolLiquidityAdded)
				if err := _FlowXPool.contract.UnpackLog(event, "LiquidityAdded", log); err != nil {
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

// ParseLiquidityAdded is a log parse operation binding the contract event 0xac1d76749e5447b7b16f5ab61447e1bd502f3bb4807af3b28e620d1700a6ee45.
//
// Solidity: event LiquidityAdded(address indexed operator, uint256 txdcAmount, uint256 tokenAmount)
func (_FlowXPool *FlowXPoolFilterer) ParseLiquidityAdded(log types.Log) (*FlowXPoolLiquidityAdded, error) {
	event := new(FlowXPoolLiquidityAdded)
	if err := _FlowXPool.contract.UnpackLog(event, "LiquidityAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlowXPoolLiquidityRemovedIterator is returned from FilterLiquidityRemoved and is used to iterate over the raw logs and unpacked data for LiquidityRemoved events raised by the FlowXPool contract.
type FlowXPoolLiquidityRemovedIterator struct {
	Event *FlowXPoolLiquidityRemoved // Event containing the contract specifics and raw log

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
func (it *FlowXPoolLiquidityRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlowXPoolLiquidityRemoved)
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
		it.Event = new(FlowXPoolLiquidityRemoved)
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
func (it *FlowXPoolLiquidityRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlowXPoolLiquidityRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlowXPoolLiquidityRemoved represents a LiquidityRemoved event raised by the FlowXPool contract.
type FlowXPoolLiquidityRemoved struct {
	Operator    common.Address
	TxdcAmount  *big.Int
	TokenAmount *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterLiquidityRemoved is a free log retrieval operation binding the contract event 0x96cd817c6329656790ef8fba7675405193677d39619571282f5e21f3a98cd059.
//
// Solidity: event LiquidityRemoved(address indexed operator, uint256 txdcAmount, uint256 tokenAmount)
func (_FlowXPool *FlowXPoolFilterer) FilterLiquidityRemoved(opts *bind.FilterOpts, operator []common.Address) (*FlowXPoolLiquidityRemovedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _FlowXPool.contract.FilterLogs(opts, "LiquidityRemoved", operatorRule)
	if err != nil {
		return nil, err
	}
	return &FlowXPoolLiquidityRemovedIterator{contract: _FlowXPool.contract, event: "LiquidityRemoved", logs: logs, sub: sub}, nil
}

// WatchLiquidityRemoved is a free log subscription operation binding the contract event 0x96cd817c6329656790ef8fba7675405193677d39619571282f5e21f3a98cd059.
//
// Solidity: event LiquidityRemoved(address indexed operator, uint256 txdcAmount, uint256 tokenAmount)
func (_FlowXPool *FlowXPoolFilterer) WatchLiquidityRemoved(opts *bind.WatchOpts, sink chan<- *FlowXPoolLiquidityRemoved, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _FlowXPool.contract.WatchLogs(opts, "LiquidityRemoved", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlowXPoolLiquidityRemoved)
				if err := _FlowXPool.contract.UnpackLog(event, "LiquidityRemoved", log); err != nil {
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

// ParseLiquidityRemoved is a log parse operation binding the contract event 0x96cd817c6329656790ef8fba7675405193677d39619571282f5e21f3a98cd059.
//
// Solidity: event LiquidityRemoved(address indexed operator, uint256 txdcAmount, uint256 tokenAmount)
func (_FlowXPool *FlowXPoolFilterer) ParseLiquidityRemoved(log types.Log) (*FlowXPoolLiquidityRemoved, error) {
	event := new(FlowXPoolLiquidityRemoved)
	if err := _FlowXPool.contract.UnpackLog(event, "LiquidityRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlowXPoolSwapIterator is returned from FilterSwap and is used to iterate over the raw logs and unpacked data for Swap events raised by the FlowXPool contract.
type FlowXPoolSwapIterator struct {
	Event *FlowXPoolSwap // Event containing the contract specifics and raw log

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
func (it *FlowXPoolSwapIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlowXPoolSwap)
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
		it.Event = new(FlowXPoolSwap)
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
func (it *FlowXPoolSwapIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlowXPoolSwapIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlowXPoolSwap represents a Swap event raised by the FlowXPool contract.
type FlowXPoolSwap struct {
	Sender   common.Address
	To       common.Address
	TxdcIn   *big.Int
	TxdcOut  *big.Int
	TokenIn  *big.Int
	TokenOut *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSwap is a free log retrieval operation binding the contract event 0xb3e2773606abfd36b5bd91394b3a54d1398336c65005baf7bf7a05efeffaf75b.
//
// Solidity: event Swap(address indexed sender, address indexed to, uint256 txdcIn, uint256 txdcOut, uint256 tokenIn, uint256 tokenOut)
func (_FlowXPool *FlowXPoolFilterer) FilterSwap(opts *bind.FilterOpts, sender []common.Address, to []common.Address) (*FlowXPoolSwapIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _FlowXPool.contract.FilterLogs(opts, "Swap", senderRule, toRule)
	if err != nil {
		return nil, err
	}
	return &FlowXPoolSwapIterator{contract: _FlowXPool.contract, event: "Swap", logs: logs, sub: sub}, nil
}

// WatchSwap is a free log subscription operation binding the contract event 0xb3e2773606abfd36b5bd91394b3a54d1398336c65005baf7bf7a05efeffaf75b.
//
// Solidity: event Swap(address indexed sender, address indexed to, uint256 txdcIn, uint256 txdcOut, uint256 tokenIn, uint256 tokenOut)
func (_FlowXPool *FlowXPoolFilterer) WatchSwap(opts *bind.WatchOpts, sink chan<- *FlowXPoolSwap, sender []common.Address, to []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _FlowXPool.contract.WatchLogs(opts, "Swap", senderRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlowXPoolSwap)
				if err := _FlowXPool.contract.UnpackLog(event, "Swap", log); err != nil {
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

// ParseSwap is a log parse operation binding the contract event 0xb3e2773606abfd36b5bd91394b3a54d1398336c65005baf7bf7a05efeffaf75b.
//
// Solidity: event Swap(address indexed sender, address indexed to, uint256 txdcIn, uint256 txdcOut, uint256 tokenIn, uint256 tokenOut)
func (_FlowXPool *FlowXPoolFilterer) ParseSwap(log types.Log) (*FlowXPoolSwap, error) {
	event := new(FlowXPoolSwap)
	if err := _FlowXPool.contract.UnpackLog(event, "Swap", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

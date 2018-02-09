package evmscc

import (
	"encoding/hex"
	"fmt"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"golang.org/x/crypto/sha3"

	pb "github.com/hyperledger/fabric/protos/peer"
)

type EvmChaincode struct{}

var logger = flogging.MustGetLogger("evmscc")
var evmLogger, _ = lifecycle.NewStdErrLogger()

func (evmscc *EvmChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debugf("Init evmscc, it's a no-op")
	return shim.Success(nil)
}

func (evmscc *EvmChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	// We always expect 2 args: Contract Address, Input/Code
	args := stub.GetArgs()
	if len(args) != 2 {
		return shim.Error(fmt.Sprintf("Expected 2 Args, Received %d", len(args)))
	}

	input, err := hex.DecodeString(string(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Recieved error when trying to decode input args: %+v", err))
	}
	statemanager := NewStateManager(stub)

	caller, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Recieved an error getting creator: %+v", err))
	}

	// Address is the last 160 bits/20 bytes of the sha256 of the public key (256 bits/32bytes)
	callerAccountSha3 := sha3.Sum256(caller)
	callerAddr := callerAccountSha3[12:]

	//Update account to ensure caller account exists in the ledger
	callerAccountAddr, err := account.AddressFromBytes(callerAddr)
	if err != nil {
		return shim.Error(fmt.Sprintf("Received error generating address for creator: %+v", err))
	}

	//Create Caller Account
	callerAccount := account.ConcreteAccount{Address: callerAccountAddr}.MutableAccount()
	err = statemanager.UpdateAccount(callerAccount)
	if err != nil {
		return shim.Error(fmt.Sprintf("Received error updating caller account"))
	}

	//Get Callee Account
	calleeAddr, err := account.AddressFromBytes(args[0])
	if err != nil {
		return shim.Error(fmt.Sprintf("Received error generating address for contract account: %+v", err))
	}
	code, err := stub.GetState(string(calleeAddr.Bytes()))
	if err != nil {
		return shim.Error(fmt.Sprintf("Recieved error getting callee account code: %+v", err))
	}

	calleeAcct := account.ConcreteAccount{Address: calleeAddr}.MutableAccount()

	if len(code) == 0 {
		code = input
	}

	//Create New VM
	vm := evm.NewVM(statemanager, evm.DefaultDynamicMemoryProvider, newParams(), account.ZeroAddress, nil, evmLogger)

	//Call the VM with current message params
	var gas uint64 = 100000
	output, err := vm.Call(callerAccount, calleeAcct, code, input, 0, &gas)
	if err != nil {

		return shim.Error(fmt.Sprintf("Recieved error while running the evm: %+v", err))
	}

	return shim.Success(output)
}

func newParams() evm.Params {
	return evm.Params{
		BlockHeight: 0,
		BlockHash:   binary.Zero256,
		BlockTime:   0,
		GasLimit:    0,
	}

}

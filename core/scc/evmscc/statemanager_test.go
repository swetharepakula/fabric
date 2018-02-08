/*
Copyright IBM Corp. 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package evmscc

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/fabric/common/mocks/scc"
	"github.com/hyperledger/fabric/core/aclmgmt"
	"github.com/hyperledger/fabric/core/aclmgmt/mocks"
	"github.com/hyperledger/fabric/core/aclmgmt/resources"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/common/sysccprovider"
	"github.com/hyperledger/fabric/core/container/util"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/stretchr/testify/assert"
)

var fakeChaincodeID = "ccid"
var fakeChannelID = "channelID"

func setupEnv() *shim.MockStub {
	sysccprovider.RegisterSystemChaincodeProviderFactory(&scc.MocksccProviderFactory{})

	mockACL := &mocks.MockACLProvider{}
	mockACL.Reset()
	mockACL.On("CheckACL", resources.LSCC_GETDEPSPEC, "channelID", &peer.SignedProposal{}).Return(nil)
	aclmgmt.RegisterACLProvider(mockACL)

	mockStub := shim.NewMockStub("mock", nil)
	mockStub.ChannelID = fakeChannelID
	return mockStub
}

func TestGetAccount(t *testing.T) {
	mockStub := setupEnv()

	codePkg := []byte("12345678901")

	var convCCID account.Address
	copy(convCCID[:], fakeChaincodeID[:])

	encodedCodePkg := encodeBytecode(t, string(convCCID.Bytes()), codePkg)

	mockStub.MockTransactionStart("transaction1")
	err := mockStub.PutState(string(convCCID.Bytes()), encodedCodePkg)
	assert.NoError(t, err)
	mockStub.MockTransactionEnd("transaction1")

	sm := NewStateManager(mockStub)

	expectedAcct := account.ConcreteAccount{
		Address: convCCID,
		Code:    codePkg,
	}.Account()

	acct, err := sm.GetAccount(convCCID)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedAcct, acct)
}

func TestGetAccountNoAccount(t *testing.T) {
	mockStub := setupEnv()

	var convCCID account.Address
	copy(convCCID[:], fakeChaincodeID[:])

	sm := NewStateManager(mockStub)

	expectedAcct := account.ConcreteAccount{}.Account()

	acct, err := sm.GetAccount(convCCID)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedAcct, acct)
}

func TestGetAccountErr(t *testing.T) {
	mockStub := setupEnv()

	codePkg := []byte("12345678901")

	var convCCID account.Address
	copy(convCCID[:], fakeChaincodeID[:])

	encodedCodePkg := encodeBytecode(t, "fake-name", codePkg)

	mockStub.MockTransactionStart("transaction1")
	err := mockStub.PutState(string(convCCID.Bytes()), encodedCodePkg)
	assert.NoError(t, err)
	mockStub.MockTransactionEnd("transaction1")

	sm := NewStateManager(mockStub)

	expectedAcct := account.ConcreteAccount{}.Account()

	acct, err := sm.GetAccount(convCCID)
	assert.Error(t, err)
	assert.EqualValues(t, expectedAcct, acct)
}

func TestGetStorage(t *testing.T) {
	mockStub := setupEnv()

	key := []byte("key")
	expectedVal := []byte("sampleStorage")

	var convCCID account.Address
	copy(convCCID[:], fakeChaincodeID[:])

	var convKey binary.Word256
	copy(convKey[:], key[:])
	compositeKey := string(convCCID.Bytes()) + string(convKey.Bytes())

	var convVal binary.Word256
	copy(convVal[:], expectedVal[:])

	mockStub.MockTransactionStart("transaction2")
	err := mockStub.PutState(string(compositeKey), convVal.Bytes())
	mockStub.MockTransactionEnd("transaction2")
	assert.NoError(t, err)

	sm := NewStateManager(mockStub)

	val, err := sm.GetStorage(convCCID, convKey)
	assert.NoError(t, err)
	assert.EqualValues(t, convVal, val)
}

//Not an allowed operation
// func TestUpdateAccount(t *testing.T) {
// 	mockStub := setupEnv()

// 	codePkg := []byte("chaincodecodepackage")
// 	codePkg2 := []byte("changedcodepackage")

// 	var convCCID account.Address
// 	copy(convCCID[:], fakeChaincodeID[:])
// 	mockStub.MockTransactionStart("transaction1")
// 	err := mockStub.PutState(string(convCCID.Bytes()), codePkg)
// 	mockStub.MockTransactionEnd("transaction1")
// 	assert.NoError(t, err)

// 	sm := NewStateManager(mockStub)

// 	updatedAcct := account.ConcreteAccount{
// 		Address: convCCID,
// 		Code:    codePkg2,
// 	}.Account()

// 	mockStub.MockTransactionStart("transaction2")
// 	err = sm.UpdateAccount(updatedAcct)
// 	mockStub.MockTransactionEnd("transaction2")
// 	assert.NoError(t, err)

// 	updatedCode, err := mockStub.GetState(string(convCCID.Bytes()))
// 	assert.NoError(t, err)
// 	assert.Equal(t, updatedCode, codePkg2)
// }

// func TestUpdateAccountNoPreviousAccount(t *testing.T) {
// 	mockStub := setupEnv()

// 	codePkg2 := []byte("changedcodepackage")

// 	var convCCID account.Address
// 	copy(convCCID[:], fakeChaincodeID[:])

// 	sm := NewStateManager(mockStub)

// 	updatedAcct := account.ConcreteAccount{
// 		Address: convCCID,
// 		Code:    codePkg2,
// 	}.Account()

// 	mockStub.MockTransactionStart("transaction1")
// 	err := sm.UpdateAccount(updatedAcct)
// 	mockStub.MockTransactionEnd("transaction1")
// 	assert.Error(t, err)
// }

//What happens with ledger associated with this?
// func TestRemoveAccount(t *testing.T) {
// 	mockStub := setupEnv()

// 	codePkg := []byte("chaincodecodepackage")

// 	var convCCID account.Address
// 	copy(convCCID[:], fakeChaincodeID[:])
// 	mockStub.MockTransactionStart("transaction1")
// 	err := mockStub.PutState(string(convCCID.Bytes()), codePkg)
// 	assert.NoError(t, err)
// 	mockStub.MockTransactionEnd("transaction1")

// 	sm := NewStateManager(mockStub)

// 	mockStub.MockTransactionStart("transaction2")
// 	err = sm.RemoveAccount(convCCID)
// 	mockStub.MockTransactionEnd("transaction2")
// 	assert.NoError(t, err)

// 	code, err := mockStub.GetState(string(convCCID.Bytes()))
// 	assert.NoError(t, err)
// 	assert.Empty(t, code)
// }

// func TestRemoveAccountNoAccount(t *testing.T) {
// 	mockStub := setupEnv()

// 	var convCCID account.Address
// 	copy(convCCID[:], fakeChaincodeID[:])

// 	sm := NewStateManager(mockStub)

// 	mockStub.MockTransactionStart("transaction1")
// 	err := sm.RemoveAccount(convCCID)
// 	mockStub.MockTransactionEnd("transaction1")
// 	assert.Error(t, err)
// }

func TestSetStorage(t *testing.T) {
	mockStub := setupEnv()

	key := []byte("key")
	expectedVal := []byte("sampleStorage")

	var convCCID account.Address
	copy(convCCID[:], fakeChaincodeID[:])

	var convKey binary.Word256
	copy(convKey[:], key[:])
	compositeKey := string(convCCID.Bytes()) + string(convKey.Bytes())

	var convVal binary.Word256
	copy(convVal[:], expectedVal[:])

	sm := NewStateManager(mockStub)

	mockStub.MockTransactionStart("transaction2")
	err := sm.SetStorage(convCCID, convKey, convVal)
	mockStub.MockTransactionEnd("transaction2")
	assert.NoError(t, err)

	val, err := mockStub.GetState(compositeKey)
	assert.NoError(t, err)
	assert.EqualValues(t, convVal.Bytes(), val)
}

func TestDecodeBytecode(t *testing.T) {
	data := []byte("234879")
	var convAddr account.Address
	copy(convAddr[:], []byte("data")[:])

	encodedBytecode := encodeBytecode(t, "data", data)

	decodedBytecode, err := DecodeBytecode(string(convAddr.Bytes()), encodedBytecode)
	assert.NoError(t, err)

	assert.Equal(t, data, decodedBytecode)
}

func encodeBytecode(t *testing.T, name string, data []byte) []byte {

	buf := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)

	err := util.WriteBytesToPackage(name, data, tw)
	assert.NoError(t, err)

	tw.Close()
	gw.Close()

	return buf.Bytes()
}

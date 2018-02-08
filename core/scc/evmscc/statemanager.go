package evmscc

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type StateManager interface {
	GetAccount(address account.Address) (account.Account, error)
	GetStorage(address account.Address, key binary.Word256) (binary.Word256, error)
	UpdateAccount(updatedAccount account.Account) error
	RemoveAccount(address account.Address) error
	SetStorage(address account.Address, key, value binary.Word256) error
}

type stateManager struct {
	stub shim.ChaincodeStubInterface
}

func NewStateManager(stub shim.ChaincodeStubInterface) StateManager {
	return &stateManager{stub: stub}
}

func (s *stateManager) GetAccount(address account.Address) (account.Account, error) {
	convAddr := string(address.Bytes())
	code, err := s.stub.GetState(convAddr)

	if err != nil {
		return account.ConcreteAccount{}.Account(), err
	}

	if code == nil {
		return account.ConcreteAccount{}.Account(), nil
	}

	decodedCode, err := DecodeBytecode(convAddr, code)
	if err != nil {
		return account.ConcreteAccount{}.Account(), err
	}

	return account.ConcreteAccount{
		Address: address,
		Code:    decodedCode,
	}.Account(), nil
}

func (s *stateManager) GetStorage(address account.Address, key binary.Word256) (binary.Word256, error) {

	compKey := string(address.Bytes()) + string(key.Bytes())

	val, err := s.stub.GetState(compKey)
	if err != nil {
		return binary.Word256{}, err
	}

	var convVal binary.Word256
	copy(convVal[:], val[:])

	return convVal, nil
}

func (s *stateManager) UpdateAccount(updatedAccount account.Account) error {
	return fmt.Errorf("NOT AN ALLOWED OPERATION: UpdateAccount")

	// convAddr := string(updatedAccount.Address().Bytes())
	// code, err := s.stub.GetState(convAddr)
	// if err != nil {
	// 	return err
	// }

	// if len(code) == 0 {
	// 	return errors.New("Account does not exist")
	// }

	// return s.stub.PutState(convAddr, updatedAccount.Code().Bytes())
}

func (s *stateManager) RemoveAccount(address account.Address) error {
	return fmt.Errorf("NOT AN ALLOWED OPERATION: RemoveAccount")
	// convAddr := string(address.Bytes())
	// code, err := s.stub.GetState(convAddr)
	// if err != nil {
	// 	return err
	// }

	// if len(code) == 0 {
	// 	return errors.New("Account does not exist")
	// }

	// return s.stub.DelState(convAddr)
}

func (s *stateManager) SetStorage(address account.Address, key, value binary.Word256) error {
	compKey := string(address.Bytes()) + string(key.Bytes())

	return s.stub.PutState(compKey, value.Bytes())
}

func DecodeBytecode(ccName string, compressedBytes []byte) ([]byte, error) {
	r := bytes.NewReader(compressedBytes)
	gr, _ := gzip.NewReader(r)
	// check for error
	tr := tar.NewReader(gr)

	defer gr.Close()

	var buf *bytes.Buffer

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return []byte{}, err
		}

		convHeaderName := stringToAddress(header.Name)

		if string(convHeaderName.Bytes()) != ccName {
			return []byte{}, fmt.Errorf("Name does not match. File: %s, CCName: %s", header.Name, ccName)
		}

		buf = bytes.NewBuffer(nil)

		io.Copy(buf, tr)

	}

	return buf.Bytes(), nil
}

func stringToAddress(addr string) account.Address {

	var convAddr account.Address
	copy(convAddr[:], []byte(addr)[:])

	return convAddr
}

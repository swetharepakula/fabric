package account

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ripemd160"
)

type Address binary.Word160

const AddressHexLength = 2 * binary.Word160Length

var ZeroAddress = Address{}

func AddressFromBytes(addr []byte) (address Address, err error) {
	if len(addr) != binary.Word160Length {
		err = fmt.Errorf("slice passed as address '%X' has %d bytes but should have %d bytes",
			addr, len(addr), binary.Word160Length)
		// It is caller's responsibility to check for errors. If they ignore the error we'll assume they want the
		// best-effort mapping of the bytes passed to an  address so we don't return here
	}
	copy(address[:], addr)
	return
}

func AddressFromHexString(str string) (Address, error) {
	bs, err := hex.DecodeString(str)
	if err != nil {
		return ZeroAddress, err
	}
	return AddressFromBytes(bs)
}

func MustAddressFromBytes(addr []byte) Address {
	address, err := AddressFromBytes(addr)
	if err != nil {
		panic(fmt.Errorf("error reading address from bytes that caller does not expect: %s", err))
	}
	return address
}

func AddressFromWord256(addr binary.Word256) Address {
	return Address(addr.Word160())
}

func (address Address) Word256() binary.Word256 {
	return binary.Word160(address).Word256()
}

// Copy address and return a slice onto the copy
func (address Address) Bytes() []byte {
	addressCopy := address
	return addressCopy[:]
}

func (address Address) String() string {
	return hex.EncodeUpperToString(address[:])
}

func (address *Address) UnmarshalJSON(data []byte) error {
	str := new(string)
	err := json.Unmarshal(data, str)
	if err != nil {
		return err
	}
	err = address.UnmarshalText([]byte(*str))
	if err != nil {
		return err
	}
	return nil
}

func (address Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeUpperToString(address[:]))
}

func (address *Address) UnmarshalText(text []byte) error {
	if len(text) != AddressHexLength {
		return fmt.Errorf("address hex '%s' has length %v but must have length %v to be a valid address",
			string(text), len(text), AddressHexLength)
	}
	_, err := hex.Decode(address[:], text)
	return err
}

func (address Address) MarshalText() ([]byte, error) {
	return ([]byte)(hex.EncodeUpperToString(address[:])), nil
}

func NewContractAddress(caller Address, sequence uint64) (newAddr Address) {
	temp := make([]byte, 32+8)
	copy(temp, caller[:])
	binary.PutUint64BE(temp[32:], uint64(sequence))
	hasher := ripemd160.New()
	hasher.Write(temp) // does not error
	copy(newAddr[:], hasher.Sum(nil))
	return
}

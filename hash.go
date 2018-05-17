/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */

package common

import (
	"bytes"
	"math/big"

	"github.com/kooksee/common/hexutil"
)

const (
	// HashLength the leghth of hash
	HashLength = 32
)

var EmptyHash Hash = Hash{}

// Hash the hash value generated by sha-3
type Hash [HashLength]byte

// BytesToHash converts bytes to hash value
func BytesToHash(b []byte) Hash {
	a := &Hash{}
	a.SetBytes(b)
	return *a
}

// StringToHash converts a string to the hash
func StringToHash(s string) Hash {
	return BytesToHash([]byte(s))
}

// SetBytes sets the hash to the value of b. If b is greater than len(a) it will panic
func (a *Hash) SetBytes(b []byte) {
	copy(a[:], b)
}

// Bytes returns its actual bits
func (a Hash) Bytes() []byte {
	return a[:]
}

// String returns the string representation of the hash
func (a Hash) String() string {
	return string(a[:])
}

// Equal returns a boolean value indicating whether the hash a is equal to the input hash b.
func (a Hash) Equal(b Hash) bool {
	return bytes.Equal(a[:], b[:])
}

// ToHex returns the hex form of the hash
func (a Hash) ToHex() string {
	return hexutil.BytesToHex(a[:])
}

func HexToHash(hex string) (Hash, error) {
	byte, err := hexutil.HexToBytes(hex)
	if err != nil {
		return EmptyHash, err
	}

	hash := BytesToHash(byte)
	return hash, nil
}

// IsEmpty return true if this hash is empty. Otherwise, false.
func (a Hash) IsEmpty() bool {
	return a == EmptyHash
}

// BigToHash converts a big int to Hash.
func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }

// Big converts this Hash to a big int.
func (a Hash) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

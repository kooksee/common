/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */

package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/kooksee/common/math"
	"github.com/kooksee/common"
	"github.com/kooksee/crypt"
	"golang.org/x/crypto/scrypt"
)

// Scrypt common parameter
const (
	ScryptN     = 1 << 18
	ScryptP     = 1
	scryptR     = 8
	scryptDKLen = 32
)

var (
	// ErrDecrypt error when the passphrase is not right
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
)

// EncryptKey encrypts a key using the specified scrypt parameters into a json
// passphrase -> script function -> decryption key
// decryption key + private key ->  aes-128-ctr algorithm -> encrypted private key
func EncryptKey(key *Key, auth string) ([]byte, error) {
	salt := getRandBuff(32)
	scryptKey, err := getScryptKey(salt, auth)
	if err != nil {
		return nil, err
	}

	encryptKey := scryptKey[:16]
	keyBytes := math.PaddedBigBytes(key.PrivateKey.D, 32)

	iv := getRandBuff(aes.BlockSize) // 16
	cipherText, err := aesCTRXOR(encryptKey, keyBytes, iv)
	if err != nil {
		return nil, err
	}

	mac := crypto.HashBytes(scryptKey[16:32], cipherText)
	info := cryptoInfo{
		CipherText: hex.EncodeToString(cipherText),
		CipherIV:   hex.EncodeToString(iv),
		Salt:       hex.EncodeToString(salt),
		MAC:        mac.ToHex(),
	}

	encryptedKey := encryptedKey{
		Version: Version,
		Address: key.Address.ToHex(),
		Crypto:  info,
	}

	return json.MarshalIndent(encryptedKey, "", "\t")
}

// DecryptKey decrypts a key from a json blob, returning the private key itself.
func DecryptKey(keyjson []byte, auth string) (*Key, error) {
	k := new(encryptedKey)
	if err := json.Unmarshal(keyjson, k); err != nil {
		return nil, err
	}
	keyBytes, err := doDecrypt(k, auth)
	// Handle any decryption errors and return the key
	if err != nil {
		return nil, err
	}
	key, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, err
	}

	addr, err := crypto.GetAddress(key)
	if err != nil {
		return nil, err
	}

	return &Key{
		Address:    *addr,
		PrivateKey: key,
	}, nil
}

func doDecrypt(keyProtected *encryptedKey, auth string) ([]byte, error) {
	if keyProtected.Version != Version {
		return nil, fmt.Errorf("Version not supported: %v", keyProtected.Version)
	}

	mac, err := common.HexToHash(keyProtected.Crypto.MAC)
	if err != nil {
		return nil, err
	}

	iv, err := hex.DecodeString(keyProtected.Crypto.CipherIV)
	if err != nil {
		return nil, err
	}

	cipherText, err := hex.DecodeString(keyProtected.Crypto.CipherText)
	if err != nil {
		return nil, err
	}

	salt, err := hex.DecodeString(keyProtected.Crypto.Salt)
	if err != nil {
		return nil, err
	}

	scyptKey, err := getScryptKey(salt, auth)
	if err != nil {
		return nil, err
	}

	calculatedMAC := crypto.HashBytes(scyptKey[16:32], cipherText)
	if !calculatedMAC.Equal(mac) {
		return nil, ErrDecrypt
	}

	plainText, err := aesCTRXOR(scyptKey[:16], cipherText, iv)
	if err != nil {
		return nil, err
	}

	return plainText, err
}

// use scrypt to calculate auth key
func getScryptKey(salt []byte, auth string) ([]byte, error) {
	authArray := []byte(auth)
	return scrypt.Key(authArray, salt, ScryptN, scryptR, ScryptP, scryptDKLen)
}

// AES-128 is selected due to size of encryptKey.
// when inText is plain text, the return value is cipher text
// when inText is cipher text, the return value is plain text
func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

func getRandBuff(n int) []byte {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}

	return mainBuff
}

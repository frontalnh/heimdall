// This file provides supporting function for key format and type.

package key

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"crypto/aes"
	"io"
	"crypto/rand"
	"crypto/cipher"
	"golang.org/x/crypto/scrypt"
)

// PEMToPublicKey converts PEM to public key format.
func PEMToPublicKey(data []byte, keyGenOpt KeyGenOpts) (PubKey, error) {

	if len(data) == 0 {
		return nil, errors.New("Input data should not be NIL")
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Failed to decode data")
	}

	key, err := DERToPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("Failed to convert PEM data to public key")
	}

	pub, err := MatchPublicKeyOpt(key, keyGenOpt)
	if err != nil {
		return nil, errors.New("Failed to convert the key type to matched public key")
	}

	return pub, nil

}

// PEMToPrivateKey converts PEM to private key format.
func PEMToPrivateKey(data []byte, keyGenOpt KeyGenOpts) (PriKey, error) {
	if len(data) == 0 {
		return nil, errors.New("Input data should not be NIL")
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Failed to decode data")
	}

	key, err := DERToPrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("Failed to convert PEM data to private key")
	}

	pri, err := MatchPrivateKeyOpt(key, keyGenOpt)
	if err != nil {
		return nil, errors.New("Failed to convert the key type to matched private key")
	}

	return pri, nil

}

// DERToPublicKey converts DER to public key format.
func DERToPublicKey(data []byte) (interface{}, error) {

	if len(data) == 0 {
		return nil, errors.New("Input data should not be NIL")
	}

	key, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, errors.New("Failed to Parse data")
	}

	return key, nil

}

// DERToPrivateKey converts DER to private key format.
func DERToPrivateKey(data []byte) (interface{}, error) {

	var key interface{}
	var err error

	if len(data) == 0 {
		return nil, errors.New("Input data should not be NIL")
	}

	if key, err := x509.ParsePKCS1PrivateKey(data); err == nil {
		return key, err
	}

	if key, err = x509.ParseECPrivateKey(data); err == nil {
		return key, err
	}

	return nil, errors.New("Unspported Private Key Type")

}

// MatchPublicKeyOpt converts key interface type to public key type using key generation option.
func MatchPublicKeyOpt(key interface{}, keyGenOpt KeyGenOpts) (publicKey PubKey, err error) {
	switch key.(type) {
	case *rsa.PublicKey:
		pub := &RSAPublicKey{PubKey: key.(*rsa.PublicKey), Bits: KeyGenOptsToRSABits(keyGenOpt)}
		return pub, nil
	case *ecdsa.PublicKey:
		pub := &ECDSAPublicKey{key.(*ecdsa.PublicKey)}
		return pub, nil
	default:
		return nil, errors.New("no matched key generation option")
	}
}

// MatchPrivateKeyOpt converts key interface type to private key type using key generation option.
func MatchPrivateKeyOpt(key interface{}, keyGenOpt KeyGenOpts) (privateKey PriKey, err error) {
	switch key.(type) {
	case *rsa.PrivateKey:
		pri := &RSAPrivateKey{PrivKey: key.(*rsa.PrivateKey), Bits: KeyGenOptsToRSABits(keyGenOpt)}
		return pri, nil
	case *ecdsa.PrivateKey:
		pri := &ECDSAPrivateKey{PrivKey: key.(*ecdsa.PrivateKey)}
		return pri, nil
	default:
		return nil, errors.New("no matched key generation option")
	}
}

// EncryptWithAES encrypts plaintext with key by AES encryption algorithm.
func EncryptWithAES(plaintext []byte, key []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext = make([]byte, aes.BlockSize + len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// EncryptWithAES encrypts plaintext with key by AES encryption algorithm.
func DecryptWithAES(ciphertext []byte, key []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext = make([]byte, len(ciphertext) - aes.BlockSize)
	iv := ciphertext[:aes.BlockSize]

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

	return plaintext, nil
}

// DeriveKeyFromPwd derives a key from input password.
func DeriveKeyFromPwd(pwd []byte, salt []byte, keyLen int) (dKey []byte, err error) {
	// The params N, r, p are cost parameters, and 32768, 8, 1 are recommended parameters for interactive login as of 2017.
	dKey, err = scrypt.Key(pwd, salt, 32768, 8, 1, keyLen)
	if err != nil {
		return nil, err
	}

	return dKey, err
}
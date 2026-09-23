package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"strings"
)

var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

func PKCS7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func PKCS7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	//fmt.Printf("b: %v\n", b)
	//fmt.Printf("len(b): %v\n", len(b))
	c := b[len(b)-1]
	//fmt.Printf("c: %v\n", c)
	n := int(c)
	//fmt.Printf("n: %v\n", n)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		//fmt.Printf("b[%d]: %v\n", len(b)-n+i, c)
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	//fmt.Printf("b: %v\n", b[:len(b)-n])
	//fmt.Printf("len(b): %v\n", len(b[:len(b)-n]))

	return b[:len(b)-n], nil
}

func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func Base64Decode(data string) string {
	dec, _ := base64.StdEncoding.DecodeString(data)
	return string(dec)
}

func MD5Hash(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func MD5HashRaw(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return string(h.Sum(nil))
}

func PublicKeyApple(base64urlEncodedN string, base64urlEncodedE string) *rsa.PublicKey {

	var pub rsa.PublicKey
	var decE, decN []byte
	var eInt int
	var err error

	//get the modulo
	decN, err = base64.RawURLEncoding.DecodeString(base64urlEncodedN)
	if err != nil {
		return nil
	}
	pub.N = new(big.Int)
	pub.N.SetBytes(decN)
	//get exponent
	decE, err = base64.RawURLEncoding.DecodeString(base64urlEncodedE)
	if err != nil {
		return nil
	}
	//convert the bytes into int
	for _, v := range decE {
		eInt = eInt << 8
		eInt = eInt | int(v)
	}
	pub.E = eInt

	return &pub
}

func AESEncrypt(key string, text string) (string, error) {
	/* Convert key hex  */
	cipherKey := []byte(key)
	/* Convert text */
	cipherText := []byte(text)
	/* Create new cipher */
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}
	/* Padding text */
	blockSize := block.BlockSize()
	origData, err := PKCS7Pad(cipherText, blockSize)
	if err != nil {
		return "", err
	}
	/* Create iv */
	cipherText = make([]byte, aes.BlockSize+len(origData))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	// fmt.Printf("AESEncrypt iv: %X\n", iv)
	// fmt.Printf("AESEncrypt cipherText: %X\n", cipherText)
	/* */
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(cipherText[aes.BlockSize:], origData)
	// fmt.Printf("AESEncrypt cipherText: %X\n", cipherText)
	return string(cipherText), nil
}

func AESEncryptWithIV(key string, vector string, text string) (string, error) {
	/* Convert key hex  */
	cipherKey := []byte(key)
	/* Convert text */
	cipherText := []byte(text)
	/* Create new cipher */
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}
	/* Padding text */
	blockSize := block.BlockSize()
	origData, err := PKCS7Pad(cipherText, blockSize)
	if err != nil {
		return "", err
	}
	/* Create iv */
	cipherText = make([]byte, aes.BlockSize+len(origData))
	iv, err := hex.DecodeString(vector)
	if err != nil {
		return "", err
	}
	for index, value := range iv {
		cipherText[index] = value
	}
	/* Encrypt */
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(cipherText[aes.BlockSize:], origData)
	return string(cipherText[aes.BlockSize:]), nil
}

func AESDecryptWithIV(key string, vector string, text string) (string, error) {
	/* Convert key hex  */
	cipherKey := []byte(key)
	/* Convert string vector to hex */
	iv, err := hex.DecodeString(vector)
	if err != nil {
		return "", err
	}
	/* Convert text */
	cipherText := []byte(text)
	/* Create new cipher */
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}
	/* Get IV key */
	if len(cipherText) < aes.BlockSize {
		return "", errors.New("Cipher Text too short")
	}
	if (len(cipherText) % aes.BlockSize) != 0 {
		return "", errors.New("Blocksize must be multipe of decoded message length")
	}
	/* Decrypt */
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(cipherText, cipherText)
	/* Unpadding text */
	unpadMsg, err := PKCS7Unpad(cipherText, block.BlockSize())
	if err != nil {
		return "", err
	}
	return string(unpadMsg), nil
}


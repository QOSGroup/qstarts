package slim

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/QOSGroup/qstars/slim/funcInlocal/respwrap"
	"io"
)

func AesEncrypt(keystring, text string) string {
	key := []byte(keystring)
	plaintext := []byte(text)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	result := base64.URLEncoding.EncodeToString(ciphertext)
	resp, _ := respwrap.ResponseWrapper(Cdc, result, nil)
	out := string(resp)
	return out
}

func AesDecrypt(keystring, cryptoText string) string {
	key := []byte(keystring)
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("Ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	result := fmt.Sprintf("%s", ciphertext)
	resp, _ := respwrap.ResponseWrapper(Cdc, result, nil)
	out := string(resp)
	return out
}

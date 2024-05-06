package util

import (
	"crypto/aes"
	"encoding/base64"
)

func AesDecrypt(point string, key string) string {
	encryptBytes, _ := base64.StdEncoding.DecodeString(point)
	info := AESDecryptECB(encryptBytes, []byte(key))
	return string(info)
}

func AESDecryptECB(data, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	decrypted := make([]byte, len(data))
	size := block.BlockSize()

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Decrypt(decrypted[bs:be], data[bs:be])
	}

	return PKCS5UnPadding(decrypted)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return origData
	}
	number := int(origData[length-1])
	return origData[:(length - number)]
}

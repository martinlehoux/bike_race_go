package auth

import (
	"bike_race/core"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

func encrypt(secret []byte, plainText string) string {
	block, err := aes.NewCipher(secret)
	if err != nil {
		err = core.Wrap(err, "error creating AES cipher")
		panic(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		err = core.Wrap(err, "error creating AEAD")
		panic(err)
	}
	nonce := make([]byte, aead.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		err = core.Wrap(err, "error generating nonce")
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(plainText), nil))
}

func decrypt(secret []byte, encryptedText string) (string, error) {
	encryptedBytes, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		err = core.Wrap(err, "error decoding base64")
		return "", err
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		err = core.Wrap(err, "error creating AES cipher")
		panic(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		err = core.Wrap(err, "error creating AEAD")
		panic(err)
	}
	nonceSize := aead.NonceSize()
	if len(encryptedBytes) < nonceSize {
		err = errors.New("encrypted text is too short")
		return "", err
	}
	nonce, cipherText := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]
	plainBytes, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		err = core.Wrap(err, "error decrypting")
		return "", err
	}
	return string(plainBytes), nil
}

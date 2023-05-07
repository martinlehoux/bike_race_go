package auth

import (
	"bike_race/core"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
)

func encrypt(secret []byte, plainText string) (string, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		err = core.Wrap(err, "error creating AES cipher")
		log.Fatal(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		err = core.Wrap(err, "error creating AEAD")
		log.Fatal(err)
	}
	nonce := make([]byte, aead.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		err = core.Wrap(err, "error generating nonce")
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(plainText), nil)), nil
}

func decrypt(secret []byte, encryptedText string) (string, error) {
	encryptedBytes, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		err = core.Wrap(err, "error decoding base64")
		log.Fatal(err)
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		err = core.Wrap(err, "error creating AES cipher")
		log.Fatal(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		err = core.Wrap(err, "error creating AEAD")
		log.Fatal(err)
	}
	nonceSize := aead.NonceSize()
	if len(encryptedBytes) < nonceSize {
		err = errors.New("encrypted text is too short")
		log.Fatal(err)
	}
	nonce, cipherText := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]
	plainBytes, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		err = core.Wrap(err, "error decrypting")
		log.Fatal(err)
	}
	return string(plainBytes), nil
}

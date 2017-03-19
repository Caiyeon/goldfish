package vault

import (
	"encoding/base64"
	"errors"
)

// encrypt given string with userTransitKey
func (auth AuthInfo) EncryptTransit(plaintext string) (string, error) {
	client, err := auth.Client()
	if err != nil {
		return "", err
	}

	resp, err := client.Logical().Write(
		"transit/encrypt/"+userTransitKey,
		map[string]interface{}{
			"plaintext": base64.StdEncoding.EncodeToString([]byte(plaintext)),
		})
	if err != nil {
		return "", err
	}

	cipher, ok := resp.Data["ciphertext"].(string)
	if !ok {
		return "", errors.New("Failed type assertion of response to string")
	}

	return cipher, nil
}

// decrypt given cipher with userTransitKey
func (auth AuthInfo) DecryptTransit(cipher string) (string, error) {
	client, err := auth.Client()
	if err != nil {
		return "", err
	}

	resp, err := client.Logical().Write(
		"transit/decrypt/"+userTransitKey,
		map[string]interface{}{
			"ciphertext": cipher,
		})
	if err != nil {
		return "", err
	}

	b64, ok := resp.Data["plaintext"].(string)
	if !ok {
		return "", errors.New("Failed type assertion of response to string")
	}

	rawbytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}

	return string(rawbytes), nil
}

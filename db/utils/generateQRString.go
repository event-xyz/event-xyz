package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(id, secret_key string) (string, string, error) {
	// Concatenate id, secret_key, and id for uniqueness
	rawString := fmt.Sprintf("%s:%s", id, secret_key)

	// Hash the concatenated string for better security
	hash := sha256.New()
	hash.Write([]byte(rawString))
	encodedString := base64.URLEncoding.EncodeToString(hash.Sum(nil))

	// Generate QR Code using the encoded string
	qrCode, err := qrcode.Encode(encodedString, qrcode.Medium, 256)
	if err != nil {
		return "", "", err
	}

	// Convert QR Code to a base64 string
	qrCodeBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(qrCode)

	return encodedString, qrCodeBase64, nil
}

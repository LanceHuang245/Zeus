package auth

import (
	"Zephyr/config"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Generate Ed25519 JWT
func GenerateJWT() (string, error) {
	priv, err := ParseEd25519PrivateKeyFromPEM([]byte(config.QweatherConfig.PrivateKeyPem))
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"sub": config.QweatherConfig.ProjectID,
		"iat": now - 30,
		"exp": now + 1800,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = config.QweatherConfig.KeyID
	return token.SignedString(priv)
}

// Parse Ed25519 PEM private key
func ParseEd25519PrivateKeyFromPEM(pemBytes []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("PEM decoding failed")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("not an Ed25519 private key")
	}
	return priv, nil
}

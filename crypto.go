package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

func getPrivateKey(passphrase string) ([]byte, []byte) {
	hash := sha256.New()
	hash.Write([]byte(passphrase))
	privateKeyRaw := hash.Sum(nil)

	privateKey := make([]byte, hex.EncodedLen(len(privateKeyRaw)))
	hex.Encode(privateKey, privateKeyRaw)

	return privateKey, privateKeyRaw
}

func getPublicKey(privateKey []byte) ([]byte, []byte) {
	curve := btcec.S256()
	x, y := curve.ScalarBaseMult(privateKey)

	publicKeyRaw := SerializePublicKey(ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	})

	publicKey := make([]byte, hex.EncodedLen(len(publicKeyRaw)))
	hex.Encode(publicKey, publicKeyRaw)

	return publicKey, publicKeyRaw
}

func SerializePublicKey(publicKey ecdsa.PublicKey) []byte {
	b := make([]byte, 0, 33)
	var format byte = 0x2
	if publicKey.Y.Bit(0) == 1 {
		format |= 0x1
	}
	b = append(b, format)
	return paddedAppend(32, b, publicKey.X.Bytes())
}

func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}

	return append(dst, src...)
}

func getAddress(publicKeyRaw []byte, networkVersion byte) string {
	ripeHash := ripemd160.New()
	if _, err := ripeHash.Write(publicKeyRaw[:]); err != nil {
		fmt.Println("Could not parse address: ", err)
	}
	addressBytes := ripeHash.Sum(nil)

	return base58.CheckEncode(addressBytes, networkVersion)
}

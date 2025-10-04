package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	fmt.Println("Generating 5 Ethereum accounts for Pano testnet...")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println()

	for i := 1; i <= 5; i++ {
		// Generate new private key
		privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		if err != nil {
			log.Fatal(err)
		}

		// Get private key bytes
		privateKeyBytes := crypto.FromECDSA(privateKey)
		
		// Get address
		address := crypto.PubkeyToAddress(privateKey.PublicKey)

		role := "Validator"
		if i > 3 {
			role = "User (100000 PANO)"
		}

		fmt.Printf("Account %d (%s %d):\n", i, role, i)
		fmt.Printf("  Address:     %s\n", address.Hex())
		fmt.Printf("  Private Key: 0x%s\n", hex.EncodeToString(privateKeyBytes))
		fmt.Println()
	}
}

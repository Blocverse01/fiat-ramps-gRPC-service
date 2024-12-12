package payments

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
)

type PaymentSequenceIdInput struct {
	TransactionHash string
	Prefix          string
}

func GeneratePaymentSequenceId(input PaymentSequenceIdInput) (string, error) {
	// Ensure the TransactionHash is provided
	if input.TransactionHash == "" {
		return "", fmt.Errorf("TransactionHash is required to generate a sequence ID")
	}

	// Create a hashed version of the TransactionHash for fixed-length uniqueness (optional)
	hash := sha256.Sum256([]byte(input.TransactionHash))
	hashedTransactionHash := hex.EncodeToString(hash[:])

	// Extract the first 16 and last 8 characters of the hash
	first16 := hashedTransactionHash[:16]
	last8 := hashedTransactionHash[len(hashedTransactionHash)-8:]

	// Combine the Prefix, first 16, and last 8 to create the sequence ID
	sequenceId := fmt.Sprintf("%s-%s%s", input.Prefix, first16, last8)

	return sequenceId, nil
}

var Percent = decimal.NewFromFloat(100)

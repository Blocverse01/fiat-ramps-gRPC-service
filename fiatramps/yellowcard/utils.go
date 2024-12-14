package yellowcard

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"regexp"
	"time"
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

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomString(length int) (string, error) {
	if length < 4 {
		return "", fmt.Errorf("length must be 4 or more")
	}

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

func GenerateCollectionSequenceId(prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("prefix is required to generate a sequence ID")
	}

	// Generate a random 24 character string
	randomString, err := GenerateRandomString(24)

	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	// Combine the Prefix and random string to create the sequence ID
	sequenceId := fmt.Sprintf("%s-%s", prefix, randomString)

	return sequenceId, nil
}

var Percent = decimal.NewFromFloat(100)

func TransformDateString(dateStr string) (string, error) {
	// Define regex patterns for the different date formats
	patterns := []string{
		`^\d{4}-\d{2}-\d{2}$`, // YYYY-MM-DD
		`^\d{4}/\d{2}/\d{2}$`, // YYYY/MM/DD
		`^\d{2}/\d{2}/\d{4}$`, // DD/MM/YYYY
		`^\d{2}-\d{2}-\d{4}$`, // DD-MM-YYYY
	}

	// Check if the date string is already in MM/DD/YYYY format
	if matched, _ := regexp.MatchString(`^\d{2}/\d{2}/\d{4}$`, dateStr); matched {
		return dateStr, nil
	}

	// Iterate over the patterns to find a match
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, dateStr); matched {
			var parsedDate time.Time
			var err error

			switch pattern {
			case `^\d{4}-\d{2}-\d{2}$`:
				parsedDate, err = time.Parse("2006-01-02", dateStr)
			case `^\d{4}/\d{2}/\d{2}$`:
				parsedDate, err = time.Parse("2006/01/02", dateStr)
			case `^\d{2}/\d{2}/\d{4}$`:
				parsedDate, err = time.Parse("02/01/2006", dateStr)
			case `^\d{2}-\d{2}-\d{4}$`:
				parsedDate, err = time.Parse("02-01-2006", dateStr)
			}

			if err != nil {
				return "", err
			}

			// Return the date in MM/DD/YYYY format
			return parsedDate.Format("01/02/2006"), nil
		}
	}

	return "", fmt.Errorf("invalid date format: %s", dateStr)
}

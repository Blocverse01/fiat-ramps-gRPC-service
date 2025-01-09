package auditing

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
	"os"
	"time"
)

func ConvertAuditExportsToCsv(transactions []AuditExport) string {
	// CSV Header
	csvString := "usdAmount,sequenceId,transactionFeeUsd,transactionHash,platformFeePercentage,blockchainNetwork,transactionType,createdAt\n"

	for _, transaction := range transactions {
		usdAmountString := decimal.NewFromFloat(transaction.UsdAmount).String()
		transactionFeeUsdString := decimal.NewFromFloat(transaction.TransactionFeeUsd).String()
		platformPercentageString := decimal.NewFromFloat(transaction.PlatformFeePercentage).String()
		parsedCreatedAt, err := time.Parse("2006-01-02T15:04:05Z", transaction.CreatedAt)

		if err != nil {
			panic(err)
		}

		formattedCreatedAt := parsedCreatedAt.Format("Jan 2, 2006 at 3:04pm (UTC)")

		csvString += usdAmountString + "," + transaction.SequenceId + "," + transactionFeeUsdString + "," + transaction.TransactionHash + "," + platformPercentageString + "," + transaction.BlockchainNetwork + "," + transaction.TransactionType + "," + formattedCreatedAt + "\n"
	}

	return csvString
}

func SaveCsvToFile(csvString, filename string) error {
	// Write the CSV string to a file
	err := os.WriteFile(filename, []byte(csvString), 0644)
	if err != nil {
		return fmt.Errorf("failed to write CSV to file: %v", err)
	}
	return nil
}

func GetTimeBefore(targetTime time.Time, duration time.Duration) time.Time {
	return targetTime.Add(-duration)
}

func EncryptWithAES(data, key, iv string) (string, error) {
	// Decode the hex-encoded key and IV
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return "", err
	}
	ivBytes, err := hex.DecodeString(iv)
	if err != nil {
		return "", err
	}

	// Ensure the key is 32 bytes long
	if len(keyBytes) != 32 {
		return "", fmt.Errorf("key length must be 32 bytes")
	}

	// Ensure the IV is 16 bytes long
	if len(ivBytes) != 16 {
		return "", fmt.Errorf("IV length must be 16 bytes")
	}

	// Create the AES cipher block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// Create the GCM mode
	aesGCM, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return "", err
	}

	// Encrypt the data
	ciphertext := aesGCM.Seal(nil, ivBytes, []byte(data), nil)

	// Return the result as a hex-encoded string
	return hex.EncodeToString(ciphertext), nil
}

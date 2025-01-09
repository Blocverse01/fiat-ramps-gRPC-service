package auditing

import "time"

const AuditExportDuration = 24 * time.Hour // 24 hours
const AuditDbBranch = "main"

type AuditExport struct {
	UsdAmount             float64 `json:"usdAmount"`
	SequenceId            string  `json:"sequenceId"`
	TransactionFeeUsd     float64 `json:"transactionFeeUsd"`
	TransactionHash       string  `json:"transactionHash"`
	PlatformFeePercentage float64 `json:"platformFeePercentage"`
	BlockchainNetwork     string  `json:"blockchainNetwork"`
	TransactionType       string  `json:"transactionType"`
	CreatedAt             string  `json:"createdAt"`
}

package database

import "net/http"

// RequiredUniqueIndexes must exist in Xata (apply in the Xata UI / schema).
// Application code also check-then-inserts and treats 409/422 as conflicts.
//
//	yellowcard_offramp_transaction.sequenceId
//	yellowcard_offramp_transaction.transactionHash
//	yellowcard_onramp_transaction.sequenceId
var RequiredUniqueIndexes = []string{
	"yellowcard_offramp_transaction.sequenceId",
	"yellowcard_offramp_transaction.transactionHash",
	"yellowcard_onramp_transaction.sequenceId",
}

func IsUniqueViolation(statusCode int) bool {
	return statusCode == http.StatusConflict || statusCode == http.StatusUnprocessableEntity
}

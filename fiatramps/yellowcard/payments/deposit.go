package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps/database"
)

const defaultBlockradarTable = "blockradar_deposits"

type DepositVerifyInput struct {
	TransactionHash   string
	HotWalletAddress  string
	FromWalletAddress string
	TokenAddress      string
	TokenSymbol       string
	BlockchainNetwork string
}

type VerifiedDeposit struct {
	TransactionHash string
	ToAddress       string
	FromAddress     string
	TokenAddress    string
	TokenAmount     float64
	Confirmed       bool
}

type DepositVerifier interface {
	Verify(input DepositVerifyInput) (VerifiedDeposit, error)
}

type BlockradarDepositRecord struct {
	TransactionHash string  `json:"transactionHash"`
	ToAddress       string  `json:"toAddress"`
	FromAddress     string  `json:"fromAddress"`
	TokenAddress    string  `json:"tokenAddress"`
	TokenSymbol     string  `json:"tokenSymbol"`
	Amount          float64 `json:"amount"`
	Status          string  `json:"status"`
	Network         string  `json:"network"`
}

type blockradarQueryResponse struct {
	Records []BlockradarDepositRecord `json:"records"`
}

// CompositeDepositVerifier checks a Blockradar/Xata record first, then optional EVM RPC.
type CompositeDepositVerifier struct{}

func (CompositeDepositVerifier) Verify(input DepositVerifyInput) (VerifiedDeposit, error) {
	if rec, err := lookupBlockradarDeposit(input.TransactionHash); err == nil {
		return matchDeposit(input, rec)
	}
	if rec, err := lookupEVMDeposit(input); err == nil {
		return matchDeposit(input, rec)
	}
	return VerifiedDeposit{}, fmt.Errorf("deposit %s could not be verified on-chain or against Blockradar", input.TransactionHash)
}

func lookupBlockradarDeposit(txHash string) (VerifiedDeposit, error) {
	table := os.Getenv("BLOCKRADAR_DEPOSITS_TABLE")
	if table == "" {
		table = defaultBlockradarTable
	}
	reqBody := database.TableDataQueryRequestBody{
		Columns: []string{"*"},
		Filter:  map[string]string{"transactionHash": txHash},
	}
	bodyData, err := fiatramps.BufferBodyData(reqBody)
	if err != nil {
		return VerifiedDeposit{}, err
	}
	resp, err := database.MakeTablesRequest("POST", fmt.Sprintf("%s/query", table), bodyData)
	if err != nil {
		return VerifiedDeposit{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return VerifiedDeposit{}, fmt.Errorf("blockradar lookup status %s", resp.Status)
	}
	var response blockradarQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return VerifiedDeposit{}, err
	}
	if len(response.Records) == 0 {
		return VerifiedDeposit{}, fmt.Errorf("no blockradar record for %s", txHash)
	}
	rec := response.Records[0]
	status := strings.ToLower(rec.Status)
	confirmed := status == "" || status == "success" || status == "confirmed" || status == "complete"
	return VerifiedDeposit{
		TransactionHash: rec.TransactionHash,
		ToAddress:       rec.ToAddress,
		FromAddress:     rec.FromAddress,
		TokenAddress:    rec.TokenAddress,
		TokenAmount:     rec.Amount,
		Confirmed:       confirmed,
	}, nil
}

type evmRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type evmTxResult struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Input string `json:"input"`
}

type evmRPCTxResponse struct {
	Result *evmTxResult `json:"result"`
}

type evmReceiptResult struct {
	Status string `json:"status"`
	To     string `json:"to"`
}

type evmRPCReceiptResponse struct {
	Result *evmReceiptResult `json:"result"`
}

func lookupEVMDeposit(input DepositVerifyInput) (VerifiedDeposit, error) {
	rpcURL := os.Getenv("EVM_RPC_URL")
	if rpcURL == "" {
		return VerifiedDeposit{}, fmt.Errorf("EVM_RPC_URL not set")
	}
	tx, err := evmRPC(rpcURL, "eth_getTransactionByHash", []interface{}{input.TransactionHash})
	if err != nil {
		return VerifiedDeposit{}, err
	}
	var txResp evmRPCTxResponse
	if err := json.Unmarshal(tx, &txResp); err != nil || txResp.Result == nil {
		return VerifiedDeposit{}, fmt.Errorf("transaction not found")
	}
	receiptRaw, err := evmRPC(rpcURL, "eth_getTransactionReceipt", []interface{}{input.TransactionHash})
	if err != nil {
		return VerifiedDeposit{}, err
	}
	var receipt evmRPCReceiptResponse
	if err := json.Unmarshal(receiptRaw, &receipt); err != nil || receipt.Result == nil {
		return VerifiedDeposit{}, fmt.Errorf("receipt not found")
	}
	if receipt.Result.Status != "0x1" {
		return VerifiedDeposit{}, fmt.Errorf("transaction not successful")
	}
	return VerifiedDeposit{
		TransactionHash: input.TransactionHash,
		ToAddress:       input.HotWalletAddress,
		FromAddress:     txResp.Result.From,
		TokenAddress:    input.TokenAddress,
		TokenAmount:     0,
		Confirmed:       true,
	}, nil
}

func evmRPC(rpcURL, method string, params []interface{}) ([]byte, error) {
	payload, err := json.Marshal(evmRPCRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params})
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(rpcURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func matchDeposit(input DepositVerifyInput, rec VerifiedDeposit) (VerifiedDeposit, error) {
	if !rec.Confirmed {
		return VerifiedDeposit{}, fmt.Errorf("deposit %s is not confirmed", input.TransactionHash)
	}
	if rec.TokenAmount <= 0 {
		return VerifiedDeposit{}, fmt.Errorf("verified deposit amount must be greater than zero")
	}
	if rec.ToAddress != "" && !equalFoldAddr(rec.ToAddress, input.HotWalletAddress) {
		return VerifiedDeposit{}, fmt.Errorf("deposit is not to the hot wallet")
	}
	if rec.FromAddress != "" && input.FromWalletAddress != "" && !equalFoldAddr(rec.FromAddress, input.FromWalletAddress) {
		return VerifiedDeposit{}, fmt.Errorf("deposit sender does not match fromWalletAddress")
	}
	if rec.TokenAddress != "" && input.TokenAddress != "" && !equalFoldAddr(rec.TokenAddress, input.TokenAddress) {
		return VerifiedDeposit{}, fmt.Errorf("deposit token does not match")
	}
	if rec.TransactionHash == "" {
		rec.TransactionHash = input.TransactionHash
	}
	return rec, nil
}

func equalFoldAddr(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

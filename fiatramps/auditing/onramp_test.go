package auditing

import (
	"fmt"
	"github.com/blocverse01/fiat-ramps-grpc-service/fiatramps"
	"testing"
	"time"
)

func TestExportOnrampAuditData(t *testing.T) {
	fiatramps.LoadEnv("../../.env")

	t.Run("Test Export Onramp Audit Data", func(t *testing.T) {
		currentTime := time.Now()
		fromTime := GetTimeBefore(currentTime, AuditExportDuration)

		auditData := ExportOnrampAuditData(fromTime, currentTime)

		fmt.Print(auditData)
	})

	t.Run("Test Generate Multiple Date Ranges", func(t *testing.T) {
		generateMultipleOnrampDateRanges()
	})
}

func generateMultipleOnrampDateRanges() {
	lowerTimeBound := time.Date(2024, 12, 23, 16, 0, 0, 0, time.UTC)
	upperTimeBound := time.Date(2025, 01, 9, 16, 0, 0, 0, time.UTC)
	currentTime := lowerTimeBound

	// Get transactions for daily range between lowerTimeBound and upperTimeBound
	for currentTime.Before(upperTimeBound) {
		nextTime := currentTime.AddDate(0, 0, 1)
		auditData := ExportOnrampAuditData(currentTime, nextTime)
		csvString := ConvertAuditExportsToCsv(auditData)

		// Save csvString to file
		fileName := fmt.Sprintf("%s---->%s_onramp_audit.csv", currentTime.Format("2006-01-02"), nextTime.Format("2006-01-02"))
		err := SaveCsvToFile(csvString, fileName)
		if err != nil {
			fmt.Println(err)
		}

		currentTime = nextTime
	}
}

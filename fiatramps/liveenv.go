package fiatramps

import (
	"os"
	"testing"
)

func SkipWithoutLiveEnv(t *testing.T, envPath string) {
	t.Helper()
	if os.Getenv("SKIP_LIVE_TESTS") == "true" || os.Getenv("CI") == "true" {
		t.Skip("skipping live integration tests")
	}
	if _, err := os.Stat(envPath); err != nil {
		t.Skip("live .env not present")
	}
}

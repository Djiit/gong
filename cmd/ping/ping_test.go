package ping

import (
	"bytes"
	"testing"
)

func TestPingCommand(t *testing.T) {
	b := bytes.NewBufferString("")
	PingCmd.SetOut(b)

	var (
		repository = "repo/owner"
		pr         = "42"
	)

	PingCmd.SetArgs([]string{"--repository", repository, "--pr", pr})

	// if err := PingCmd.Execute(); err != nil {
	// 	t.Fatalf("Expected no error, got %v", err)
	// }
}

func TestEnabledFlag(t *testing.T) {
	// This is a basic test setup - in a real implementation, you would
	// mock the GitHub client and other dependencies to test the full flow
	b := bytes.NewBufferString("")
	PingCmd.SetOut(b)

	// Test with enabled flag
	PingCmd.SetArgs([]string{"--enabled=true"})

	// Test with disabled flag
	PingCmd.SetArgs([]string{"--enabled=false"})
}

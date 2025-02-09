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

package grpc

import (
	"context"
	"testing"
	"time"
)

func TestValidateTimeSettings(t *testing.T) {
	valid := TimeSettings{
		Short:   time.Second,
		Default: 2 * time.Second,
		Long:    3 * time.Second,
	}
	if err := validateTimeSettings(valid); err != nil {
		t.Fatalf("validateTimeSettings() error = %v", err)
	}

	invalid := []TimeSettings{
		{},
		{Short: time.Second, Default: 2 * time.Second},
		{Short: 3 * time.Second, Default: 2 * time.Second, Long: 4 * time.Second},
		{Short: time.Second, Default: 4 * time.Second, Long: 3 * time.Second},
	}
	for _, settings := range invalid {
		if err := validateTimeSettings(settings); err == nil {
			t.Fatalf("validateTimeSettings(%+v) error = nil", settings)
		}
	}
}

func TestRPCContextUsesConfiguredTimeout(t *testing.T) {
	client := &Client{
		rpcTimeout: TimeSettings{
			Short:   time.Second,
			Default: 2 * time.Second,
			Long:    3 * time.Second,
		},
	}

	ctx, cancel := client.rpcDefaultContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("RPC context has no deadline")
	}

	remaining := time.Until(deadline)
	if remaining <= time.Second || remaining > 2*time.Second {
		t.Fatalf("RPC context deadline remaining = %s, want approximately 2s", remaining)
	}
}

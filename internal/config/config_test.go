package config

import (
	"testing"
	"time"
)

func TestEnvDuration(t *testing.T) {
	t.Setenv("TEST_DURATION", "250ms")

	got, err := envDuration("TEST_DURATION", time.Second)
	if err != nil {
		t.Fatalf("envDuration() error = %v", err)
	}
	if got != 250*time.Millisecond {
		t.Fatalf("envDuration() = %s, want %s", got, 250*time.Millisecond)
	}
}

func TestEnvDurationRejectsInvalidValue(t *testing.T) {
	for _, value := range []string{"invalid", "0s", "-1s"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("TEST_DURATION", value)

			if _, err := envDuration("TEST_DURATION", time.Second); err == nil {
				t.Fatalf("envDuration(%q) error = nil", value)
			}
		})
	}
}

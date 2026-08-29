package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadModel_ReturnsModelFromSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"model": "us.anthropic.claude-opus-5", "env": {"FOO": "bar"}}`), 0644)

	got, err := readModel(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "us.anthropic.claude-opus-5" {
		t.Errorf("got %q, want %q", got, "us.anthropic.claude-opus-5")
	}
}

func TestReadModel_ReturnsEmptyWhenNoModelKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"env": {}}`), 0644)

	got, err := readModel(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestWriteModel_UpdatesModelPreservingOtherFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"model": "old-model", "env": {"KEY": "value"}}`), 0644)

	err := writeModel(path, "us.anthropic.claude-sonnet-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := readModel(path)
	if got != "us.anthropic.claude-sonnet-5" {
		t.Errorf("model = %q, want %q", got, "us.anthropic.claude-sonnet-5")
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"KEY"`) {
		t.Error("other fields were not preserved")
	}
}

func TestWriteModel_FailsWhenFileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	err := writeModel(path, "anything")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

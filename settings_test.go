package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestWriteSettings_UpdatesModelPreservingOtherFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"model": "old-model", "env": {"KEY": "value"}}`), 0644)

	err := writeSettings(path, "us.anthropic.claude-sonnet-5", awsEnv{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := readModel(path)
	if got != "us.anthropic.claude-sonnet-5" {
		t.Errorf("model = %q, want %q", got, "us.anthropic.claude-sonnet-5")
	}

	if readEnv(t, path)["KEY"] != "value" {
		t.Error("other env fields were not preserved")
	}
}

func TestWriteSettings_FailsWhenFileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	err := writeSettings(path, "anything", awsEnv{})
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestWriteSettings_CreatesEnvWhenAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"model": "old-model"}`), 0644)

	if err := writeSettings(path, "m", awsEnv{Profile: "my-profile", Region: "eu-west-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env := readEnv(t, path)
	if env["CLAUDE_CODE_USE_BEDROCK"] != "1" {
		t.Errorf("CLAUDE_CODE_USE_BEDROCK = %v, want \"1\"", env["CLAUDE_CODE_USE_BEDROCK"])
	}
	if env["AWS_PROFILE"] != "my-profile" {
		t.Errorf("AWS_PROFILE = %v, want %q", env["AWS_PROFILE"], "my-profile")
	}
	if env["AWS_REGION"] != "eu-west-1" {
		t.Errorf("AWS_REGION = %v, want %q", env["AWS_REGION"], "eu-west-1")
	}
}

func TestWriteSettings_OverwritesExistingAWSValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"env": {"AWS_PROFILE": "old", "AWS_REGION": "us-east-1", "KEY": "keep"}}`), 0644)

	if err := writeSettings(path, "m", awsEnv{Profile: "new-profile", Region: "ap-southeast-2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env := readEnv(t, path)
	if env["AWS_PROFILE"] != "new-profile" {
		t.Errorf("AWS_PROFILE = %v, want %q", env["AWS_PROFILE"], "new-profile")
	}
	if env["AWS_REGION"] != "ap-southeast-2" {
		t.Errorf("AWS_REGION = %v, want %q", env["AWS_REGION"], "ap-southeast-2")
	}
	if env["KEY"] != "keep" {
		t.Error("unrelated env key was not preserved")
	}
}

func TestWriteSettings_SkipsEmptyProfileAndRegion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"env": {"AWS_PROFILE": "keep", "AWS_REGION": "keep-region"}}`), 0644)

	if err := writeSettings(path, "m", awsEnv{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env := readEnv(t, path)
	if env["AWS_PROFILE"] != "keep" {
		t.Errorf("AWS_PROFILE = %v, want it left untouched", env["AWS_PROFILE"])
	}
	if env["AWS_REGION"] != "keep-region" {
		t.Errorf("AWS_REGION = %v, want it left untouched", env["AWS_REGION"])
	}
	if env["CLAUDE_CODE_USE_BEDROCK"] != "1" {
		t.Errorf("CLAUDE_CODE_USE_BEDROCK = %v, want \"1\"", env["CLAUDE_CODE_USE_BEDROCK"])
	}
}

func readEnv(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parsing settings: %v", err)
	}
	env, _ := settings["env"].(map[string]any)
	return env
}

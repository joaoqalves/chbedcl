package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListAWSProfiles_ParsesProfileSections(t *testing.T) {
	dir := t.TempDir()
	awsDir := filepath.Join(dir, ".aws")
	os.MkdirAll(awsDir, 0755)
	os.WriteFile(filepath.Join(awsDir, "config"), []byte(`[default]
region = us-east-1

[profile staging]
region = eu-west-1

[profile production]
region = us-west-2
`), 0644)

	t.Setenv("HOME", dir)

	profiles, err := listAWSProfiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"default", "staging", "production"}
	if len(profiles) != len(want) {
		t.Fatalf("got %d profiles, want %d: %v", len(profiles), len(want), profiles)
	}
	for i := range want {
		if profiles[i] != want[i] {
			t.Errorf("profiles[%d] = %q, want %q", i, profiles[i], want[i])
		}
	}
}

func TestProfileRegion_ReturnsRegionForProfile(t *testing.T) {
	dir := t.TempDir()
	awsDir := filepath.Join(dir, ".aws")
	os.MkdirAll(awsDir, 0755)
	os.WriteFile(filepath.Join(awsDir, "config"), []byte(`[default]
region = us-east-1

[profile staging]
region = eu-west-1
output = json
`), 0644)

	t.Setenv("HOME", dir)

	if got := profileRegion("staging"); got != "eu-west-1" {
		t.Errorf("profileRegion(staging) = %q, want %q", got, "eu-west-1")
	}
	if got := profileRegion("default"); got != "us-east-1" {
		t.Errorf("profileRegion(default) = %q, want %q", got, "us-east-1")
	}
	if got := profileRegion("nonexistent"); got != "" {
		t.Errorf("profileRegion(nonexistent) = %q, want empty", got)
	}
}

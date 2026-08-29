package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadCache_ReturnsFalseWhenFileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	_, ok := loadCacheFrom(path, "us-east-1", "default")
	if ok {
		t.Error("expected cache miss for missing file")
	}
}

func TestLoadCache_ReturnsFalseWhenExpired(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	c := cache{
		FetchedAt: time.Now().Add(-49 * time.Hour),
		Region:    "us-east-1",
		Profile:   "default",
		Models:    []string{"us.anthropic.claude-opus-5"},
	}
	data, _ := json.Marshal(c)
	os.WriteFile(path, data, 0644)

	_, ok := loadCacheFrom(path, "us-east-1", "default")
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestLoadCache_ReturnsFalseWhenRegionMismatches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	c := cache{
		FetchedAt: time.Now(),
		Region:    "us-east-1",
		Profile:   "default",
		Models:    []string{"us.anthropic.claude-opus-5"},
	}
	data, _ := json.Marshal(c)
	os.WriteFile(path, data, 0644)

	_, ok := loadCacheFrom(path, "eu-west-1", "default")
	if ok {
		t.Error("expected cache miss when region differs")
	}
}

func TestLoadCache_ReturnsFalseWhenProfileMismatches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	c := cache{
		FetchedAt: time.Now(),
		Region:    "us-east-1",
		Profile:   "prod",
		Models:    []string{"us.anthropic.claude-opus-5"},
	}
	data, _ := json.Marshal(c)
	os.WriteFile(path, data, 0644)

	_, ok := loadCacheFrom(path, "us-east-1", "dev")
	if ok {
		t.Error("expected cache miss when profile differs")
	}
}

func TestSaveAndLoadCache_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	want := []string{"us.anthropic.claude-opus-5", "us.anthropic.claude-sonnet-5"}

	err := saveCacheTo(path, "us-east-1", "myprofile", want)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, ok := loadCacheFrom(path, "us-east-1", "myprofile")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != len(want) {
		t.Fatalf("got %d models, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("model[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLoadCache_UsesResolvedRegionAsKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	models := []string{"us.anthropic.claude-opus-5"}

	saveCacheTo(path, "us-east-1", "default", models)

	_, ok := loadCacheFrom(path, "us-east-1", "default")
	if !ok {
		t.Error("expected cache hit when region matches resolved value")
	}

	_, ok = loadCacheFrom(path, "", "default")
	if ok {
		t.Error("expected cache miss when raw empty region doesn't match saved resolved region")
	}
}

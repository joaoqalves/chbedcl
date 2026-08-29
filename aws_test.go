package main

import "testing"

func TestRegionPrefix(t *testing.T) {
	tests := []struct {
		region string
		want   string
	}{
		{"us-east-1", "us"},
		{"us-west-2", "us"},
		{"eu-west-1", "eu"},
		{"eu-central-1", "eu"},
		{"ap-northeast-1", "ap"},
		{"ap-southeast-2", "ap"},
		{"", "us"},
	}

	for _, tt := range tests {
		if got := regionPrefix(tt.region); got != tt.want {
			t.Errorf("regionPrefix(%q) = %q, want %q", tt.region, got, tt.want)
		}
	}
}

func TestParseFoundationModels_FiltersLegacyAndNonInferenceProfile(t *testing.T) {
	models := []foundationModel{
		{id: "anthropic.claude-opus-5", status: "ACTIVE", supportsInferenceProfile: true},
		{id: "anthropic.claude-old", status: "LEGACY", supportsInferenceProfile: true},
		{id: "anthropic.claude-provisioned", status: "ACTIVE", supportsInferenceProfile: false},
	}

	got := parseFoundationModels(models, "us-east-1")

	if len(got) != 1 {
		t.Fatalf("got %d models, want 1", len(got))
	}
	if got[0] != "us.anthropic.claude-opus-5" {
		t.Errorf("got %q, want %q", got[0], "us.anthropic.claude-opus-5")
	}
}

func TestParseFoundationModels_AppliesRegionPrefix(t *testing.T) {
	models := []foundationModel{
		{id: "anthropic.claude-sonnet-5", status: "ACTIVE", supportsInferenceProfile: true},
	}

	got := parseFoundationModels(models, "eu-west-1")

	if len(got) != 1 || got[0] != "eu.anthropic.claude-sonnet-5" {
		t.Errorf("got %v, want [eu.anthropic.claude-sonnet-5]", got)
	}
}

func TestParseFoundationModels_EmptyList(t *testing.T) {
	got := parseFoundationModels(nil, "us-east-1")
	if len(got) != 0 {
		t.Errorf("got %d models, want 0", len(got))
	}
}

func TestParseFoundationModels_SortsAlphabetically(t *testing.T) {
	models := []foundationModel{
		{id: "anthropic.claude-sonnet-5", status: "ACTIVE", supportsInferenceProfile: true},
		{id: "anthropic.claude-haiku-4", status: "ACTIVE", supportsInferenceProfile: true},
		{id: "anthropic.claude-opus-5", status: "ACTIVE", supportsInferenceProfile: true},
	}

	got := parseFoundationModels(models, "us-east-1")

	want := []string{
		"us.anthropic.claude-haiku-4",
		"us.anthropic.claude-opus-5",
		"us.anthropic.claude-sonnet-5",
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

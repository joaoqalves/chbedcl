package main

import (
	"encoding/json"
	"os"
	"time"
)

const cacheTTL = 48 * time.Hour

type cache struct {
	FetchedAt time.Time `json:"fetched_at"`
	Region    string    `json:"region"`
	Profile   string    `json:"profile"`
	Models    []string  `json:"models"`
}

func loadCacheFrom(path, region, profile string) ([]string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var c cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, false
	}

	if time.Since(c.FetchedAt) > cacheTTL {
		return nil, false
	}
	if c.Region != region || c.Profile != profile {
		return nil, false
	}

	return c.Models, true
}

func saveCacheTo(path, region, profile string, models []string) error {
	c := cache{
		FetchedAt: time.Now(),
		Region:    region,
		Profile:   profile,
		Models:    models,
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func loadCache(region, profile string) ([]string, bool) {
	path, err := cachePath()
	if err != nil {
		return nil, false
	}
	return loadCacheFrom(path, region, profile)
}

func saveCache(region, profile string, models []string) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	return saveCacheTo(path, region, profile, models)
}

func getModels(opts Options) (modelResult, error) {
	if !opts.Refresh {
		if models, ok := loadCache(opts.Region, opts.Profile); ok {
			return modelResult{Models: models, Region: opts.Region}, nil
		}
	}

	result, err := fetchModels(opts)
	if err != nil {
		return modelResult{}, err
	}

	_ = saveCache(result.Region, opts.Profile, result.Models)
	return result, nil
}

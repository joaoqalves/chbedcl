package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Prefs struct {
	Profile string `json:"profile,omitempty"`
	Region  string `json:"region,omitempty"`
}

func prefsPath() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "chbedcl-prefs.json"), nil
}

func loadPrefs() (Prefs, error) {
	path, err := prefsPath()
	if err != nil {
		return Prefs{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Prefs{}, nil
	}

	var p Prefs
	json.Unmarshal(data, &p)
	return p, nil
}

func savePrefs(p Prefs) error {
	path, err := prefsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

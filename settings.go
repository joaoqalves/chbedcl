package main

import (
	"encoding/json"
	"os"
)

func readModel(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return "", err
	}

	model, _ := settings["model"].(string)
	return model, nil
}

func writeModel(path, model string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	settings["model"] = model

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(out, '\n'), 0644)
}

func currentModel() (string, error) {
	path, err := settingsPath()
	if err != nil {
		return "", err
	}
	return readModel(path)
}

func updateModel(model string) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	return writeModel(path, model)
}

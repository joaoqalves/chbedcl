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

// awsEnv holds the AWS credentials context written into the settings.json
// "env" block so Claude Code targets the same profile/region chbedcl used.
type awsEnv struct {
	Profile string
	Region  string
}

func writeSettings(path, model string, aws awsEnv) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	settings["model"] = model

	env, _ := settings["env"].(map[string]any)
	if env == nil {
		env = map[string]any{}
	}
	env["CLAUDE_CODE_USE_BEDROCK"] = "1"
	if aws.Profile != "" {
		env["AWS_PROFILE"] = aws.Profile
	}
	if aws.Region != "" {
		env["AWS_REGION"] = aws.Region
	}
	settings["env"] = env

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

func updateSettings(model string, aws awsEnv) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	return writeSettings(path, model, aws)
}

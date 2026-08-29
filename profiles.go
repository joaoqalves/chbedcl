package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func listAWSProfiles() ([]string, error) {
	home, err := homeDir()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filepath.Join(home, ".aws", "config"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var profiles []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
			name := line[len("[profile ") : len(line)-1]
			profiles = append(profiles, name)
		} else if line == "[default]" {
			profiles = append(profiles, "default")
		}
	}

	return profiles, scanner.Err()
}

func profileRegion(profile string) string {
	home, err := homeDir()
	if err != nil {
		return ""
	}

	f, err := os.Open(filepath.Join(home, ".aws", "config"))
	if err != nil {
		return ""
	}
	defer f.Close()

	inProfile := false
	target := "[profile " + profile + "]"
	if profile == "default" {
		target = "[default]"
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == target {
			inProfile = true
			continue
		}
		if inProfile {
			if strings.HasPrefix(line, "[") {
				break
			}
			if strings.HasPrefix(line, "region") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	return ""
}

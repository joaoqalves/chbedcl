package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Options struct {
	Region           string
	Profile          string
	Refresh          bool
	JSON             bool
	OverrideDefaults bool
}

// promptMode controls whether resolveOpts may fall back to interactive TUI
// prompts (profile picker, "save as default?") when options are ambiguous.
type promptMode int

const (
	allowPrompts promptMode = iota
	noPrompts
)

// promptModeFor picks the prompt mode for a run: JSON/headless output never
// prompts, interactive output may.
func promptModeFor(jsonOut bool) promptMode {
	if jsonOut {
		return noPrompts
	}
	return allowPrompts
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `chbedcl — change the Claude model in ~/.claude/settings.json

Usage:
  chbedcl                       Interactive model picker (TUI)
  chbedcl --list                List available models
  chbedcl --set <model>         Set model directly
  chbedcl --current             Show current model

AWS options:
  -p, --profile <name>          AWS profile (overrides saved default)
  -r, --region <region>         AWS region (overrides saved default)
      --override-defaults       Save --profile/--region as new defaults (headless)
      --refresh                 Bypass 48h model cache

Output options:
      --json                    Output as JSON (for scripts/agents)

On first run, chbedcl prompts you to select an AWS profile and remembers
your choice. Use --profile/--region to override, and --override-defaults
to update the saved preference without a prompt.

Setting a model also syncs the "env" block in settings.json so that
AWS_PROFILE and AWS_REGION match the resolved profile/region and
CLAUDE_CODE_USE_BEDROCK is set to "1".

Examples:
  chbedcl                                         Pick a model interactively
  chbedcl --list --json                           List models as JSON
  chbedcl --set us.anthropic.claude-opus-5        Set model headlessly
  chbedcl -p my-profile -r eu-west-1             Use a different profile/region
  chbedcl -p new-profile --override-defaults     Change saved defaults
`)
	}

	var opts Options
	flag.StringVar(&opts.Region, "region", "", "AWS region (overrides saved default)")
	flag.StringVar(&opts.Region, "r", "", "AWS region (shorthand)")
	flag.StringVar(&opts.Profile, "profile", "", "AWS profile (overrides saved default)")
	flag.StringVar(&opts.Profile, "p", "", "AWS profile (shorthand)")
	flag.BoolVar(&opts.Refresh, "refresh", false, "force cache refresh")
	flag.BoolVar(&opts.JSON, "json", false, "output as JSON")
	flag.BoolVar(&opts.OverrideDefaults, "override-defaults", false, "save --profile/--region as new defaults (headless)")

	listFlag := flag.Bool("list", false, "list available models and exit")
	setFlag := flag.String("set", "", "set model directly (no TUI)")
	currentFlag := flag.Bool("current", false, "print current model and exit")
	flag.Parse()

	if *currentFlag {
		cur, err := currentModel()
		if err != nil {
			fatal(opts.JSON, "reading settings: %v", err)
		}
		if opts.JSON {
			printJSON(map[string]string{"model": cur})
		} else {
			fmt.Println(cur)
		}
		return
	}

	if *setFlag != "" {
		if err := resolveOpts(&opts, noPrompts); err != nil {
			fatal(opts.JSON, "%v", err)
		}
		if err := updateSettings(*setFlag, awsEnv{Profile: opts.Profile, Region: opts.Region}); err != nil {
			fatal(opts.JSON, "updating settings: %v", err)
		}
		if opts.JSON {
			printJSON(map[string]string{"model": *setFlag, "status": "ok"})
		} else {
			fmt.Printf("Model set to: %s\n", *setFlag)
		}
		return
	}

	if err := resolveOpts(&opts, promptModeFor(opts.JSON)); err != nil {
		fatal(opts.JSON, "%v", err)
	}

	result, err := getModels(opts)
	if err != nil {
		fatal(opts.JSON, "fetching models: %v", err)
	}

	if len(result.Models) == 0 {
		fatal(opts.JSON, "no Claude models found")
	}

	if *listFlag {
		if opts.JSON {
			cur, _ := currentModel()
			printJSON(map[string]any{"models": result.Models, "current": cur})
		} else {
			for _, m := range result.Models {
				fmt.Println(m)
			}
		}
		return
	}

	cur, _ := currentModel()
	selected, err := pickModel(result.Models, cur)
	if err != nil {
		fatal(opts.JSON, "%v", err)
	}
	if selected == "" {
		return
	}

	if err := updateSettings(selected, awsEnv{Profile: opts.Profile, Region: result.Region}); err != nil {
		fatal(opts.JSON, "updating settings: %v", err)
	}
	fmt.Printf("Model set to: %s\n", selected)
}

func resolveOpts(opts *Options, mode promptMode) error {
	prefs, _ := loadPrefs()
	flagsProvided := opts.Profile != "" || opts.Region != ""

	if opts.Profile == "" {
		if prefs.Profile != "" {
			opts.Profile = prefs.Profile
		} else if mode == allowPrompts {
			profiles, err := listAWSProfiles()
			if err != nil {
				return fmt.Errorf("listing AWS profiles: %w", err)
			}
			if len(profiles) == 0 {
				return fmt.Errorf("no profiles found in ~/.aws/config")
			}
			selected, err := pickFromList("Select AWS profile", profiles)
			if err != nil {
				return err
			}
			if selected == "" {
				os.Exit(0)
			}
			opts.Profile = selected
			_ = savePrefs(Prefs{Profile: opts.Profile, Region: opts.Region})
		}
	}

	if opts.Region == "" {
		if prefs.Region != "" {
			opts.Region = prefs.Region
		} else if r := profileRegion(opts.Profile); r != "" {
			opts.Region = r
		}
	}

	if !flagsProvided {
		return nil
	}

	if opts.OverrideDefaults {
		_ = savePrefs(Prefs{Profile: opts.Profile, Region: opts.Region})
		return nil
	}

	if mode == allowPrompts && (opts.Profile != prefs.Profile || opts.Region != prefs.Region) {
		save, err := pickFromList("Save as new default?", []string{"Yes", "No"})
		if err != nil {
			return err
		}
		if save == "Yes" {
			_ = savePrefs(Prefs{Profile: opts.Profile, Region: opts.Region})
		}
	}

	return nil
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func fatal(jsonOut bool, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if jsonOut {
		printJSON(map[string]string{"error": msg})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
	os.Exit(1)
}

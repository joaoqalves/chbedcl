package main

import (
	"context"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
)

func regionPrefix(region string) string {
	switch {
	case strings.HasPrefix(region, "us-"):
		return "us"
	case strings.HasPrefix(region, "eu-"):
		return "eu"
	case strings.HasPrefix(region, "ap-"):
		return "ap"
	default:
		return "us"
	}
}

func parseFoundationModels(models []foundationModel, region string) []string {
	prefix := regionPrefix(region)
	var result []string
	for _, m := range models {
		if m.status == "LEGACY" {
			continue
		}
		if !m.supportsInferenceProfile {
			continue
		}
		result = append(result, prefix+"."+m.id)
	}
	sort.Strings(result)
	return result
}

type foundationModel struct {
	id                       string
	status                   string
	supportsInferenceProfile bool
}

type modelResult struct {
	Models []string
	Region string
}

func fetchModels(opts Options) (modelResult, error) {
	ctx := context.Background()

	var cfgOpts []func(*config.LoadOptions) error
	if opts.Region != "" {
		cfgOpts = append(cfgOpts, config.WithRegion(opts.Region))
	}
	if opts.Profile != "" {
		cfgOpts = append(cfgOpts, config.WithSharedConfigProfile(opts.Profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return modelResult{}, err
	}

	client := bedrock.NewFromConfig(cfg)
	provider := "Anthropic"
	out, err := client.ListFoundationModels(ctx, &bedrock.ListFoundationModelsInput{
		ByProvider: &provider,
	})
	if err != nil {
		return modelResult{}, err
	}

	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	var models []foundationModel
	for _, m := range out.ModelSummaries {
		sup := false
		for _, t := range m.InferenceTypesSupported {
			if string(t) == "INFERENCE_PROFILE" {
				sup = true
				break
			}
		}
		models = append(models, foundationModel{
			id:                       *m.ModelId,
			status:                   string(m.ModelLifecycle.Status),
			supportsInferenceProfile: sup,
		})
	}

	return modelResult{
		Models: parseFoundationModels(models, region),
		Region: region,
	}, nil
}

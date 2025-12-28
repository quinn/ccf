package esm

import (
	"flag"
	"fmt"
	"os"

	"go.quinn.io/ccf/esm"
	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	EsmVendor esm.Config `yaml:"esm-vendor"`
}

func Main() {
	configFile := flag.String("config", "ccf.yaml", "Config file path")
	flag.Parse()

	config, err := loadConfigFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
		os.Exit(1)
	}

	if err := esm.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func loadConfigFile(path string) (esm.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return esm.Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var configFile ConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return esm.Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return configFile.EsmVendor, nil
}

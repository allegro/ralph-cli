package main

// Configuration for ralph-cli.
type Config struct {
	Debug     bool
	LogOutput string // e.g. logstash
}

// GetConfig loads ralph-cli configuration from ~/.ralph-cli dir.
func GetConfig() (Config, error) {
	return Config{
		Debug:     false,
		LogOutput: "",
	}, nil
}

// CreateDefault creates default ralph-cli config in ~/.ralph-cli dir, if not present.
func CreateDefault() error {
	return nil
}

// CreatePythonVenv creates a virtualenv for Python scripts in ~/.ralph-cli dir.
func CreatePythonVenv() error {
	return nil
}

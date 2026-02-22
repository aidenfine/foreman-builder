package foremanbuilder

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Packages []string `yaml:"packages"`
}

func ParseConfig(r io.Reader) (Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func GetYmlValues(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	return ParseConfig(file)
}

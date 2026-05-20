package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type CheckType string

const (
	HTTP CheckType = "http"
	TCP  CheckType = "tcp"
)

type Service struct {
	Name     string        `yaml:"name"`
	Check    CheckType     `yaml:"check"`
	Target   string        `yaml:"target"`
	Interval time.Duration `yaml:"interval"`
}

func (s *Service) UnmarshalYAML(value *yaml.Node) error {
	type raw struct {
		Name     string    `yaml:"name"`
		Check    CheckType `yaml:"check"`
		Target   string    `yaml:"target"`
		Interval string    `yaml:"interval"`
	}
	var r raw

	if err := value.Decode(&r); err != nil {
		return err
	}

	d, err := time.ParseDuration(r.Interval)
	if err != nil {
		return fmt.Errorf("service %q: invalid internal %q", r.Name, r.Interval)
	}

	s.Name = r.Name
	s.Check = r.Check
	s.Target = r.Target
	s.Interval = d

	return nil
}

type Config struct {
	Services []Service `yaml:"services"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

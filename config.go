package bob

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadConfig reads a JSON configuration file from disk and returns a parsed Config struct.
func LoadConfig(path string) (*Config, error) {
	configPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to determine absolute path: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat the path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &config, nil
}

// Config contains configuration settings for a service.
type Config struct {
	Docker   *DockerConfig     `json:"docker"`
	Go       *GoConfig         `json:"go,omitempty"`
	Flags    map[string]string `json:"flags,omitempty"`
	Secrets  []string          `json:"secrets,omitempty"`
	Service  *ServiceConfig    `json:"service"`
	Requires []string          `json:"requires"`
}

func (conf *Config) Validate() error {
	if conf.Service == nil {
		return fmt.Errorf("missing required field: service")
	}

	if err := conf.Service.Validate(); err != nil {
		return fmt.Errorf("validate config.service: %w", err)
	}

	if conf.Docker == nil {
		return fmt.Errorf("missing required field: docker")
	}

	// We must validate Docker after Service in case we need to generate the image name.
	if conf.Docker.Image == "" {
		img, err := generateImageName(conf.Service.Name)
		if err != nil {
			return fmt.Errorf("generate image name: %w", err)
		}

		conf.Docker.Image = img
	}

	if conf.Go != nil {
		if err := conf.Go.Validate(); err != nil {
			return fmt.Errorf("validate config.go: %w", err)
		}
	}

	return nil
}

// DockerConfig contains configuration settings about a service's docker build.
type DockerConfig struct {
	DependsOn     []string       `json:"dependsOn,omitempty"`
	Image         string         `json:"image,omitempty"`
	RestartPolicy string         `json:"restartPolicy,omitempty"`
	Volumes       []VolumeConfig `json:"volumes,omitempty"`
}

// Validate ensures that populated fields are valid.
func (conf *DockerConfig) Validate() error {
	for _, v := range conf.Volumes {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validate volume: %w", err)
		}
	}

	return nil
}

// GoConfig contains configuration settings about a Go service's build.
type GoConfig struct {
	Arch      string `json:"arch,omitempty"`
	BinaryDir string `json:"binaryDir,omitempty"`
	Module    string `json:"module,omitempty"`
	OS        string `json:"os,omitempty"`
	Version   string `json:"version,omitempty"`
}

// Validate ensures that all required fields are populated.
func (conf *GoConfig) Validate() error {
	if conf.Arch == "" {
		return fmt.Errorf("missing required field: go.arch")
	}
	if conf.BinaryDir == "" {
		return fmt.Errorf("missing required field: go.binaryDir")
	}
	if conf.Module == "" {
		return fmt.Errorf("missing required field: go.module")
	}
	if conf.OS == "" {
		return fmt.Errorf("missing required field: go.os")
	}
	if conf.Version == "" {
		return fmt.Errorf("missing required field: go.version")
	}
	return nil
}

// ServiceConfig contains configuration settings about how to run a service.
type ServiceConfig struct {
	Name        string   `json:"name"`
	Host        string   `json:"host"`
	Expose      []int    `json:"expose,omitempty"`
	Ports       []string `json:"ports,omitempty"`
	MetricsPort int      `json:"metricsPort,omitempty"`
}

// Validate ensures that all required fields are populated.
func (conf *ServiceConfig) Validate() error {
	if conf.Name == "" {
		return fmt.Errorf("missing required field: service.name")
	}

	if conf.Host == "" {
		return fmt.Errorf("missing required field: service.host")
	}

	return nil
}

// VolumeType describes a kind of Docker volume.
type VolumeType string

const (
	VolumeTypeBind  = "bind"
	VolumeTypeNamed = "named"
)

// VolumeConfig contains configuration settings for a Docker volume.
type VolumeConfig struct {
	Source string     `json:"source"`
	Mount  string     `json:"mount"`
	Type   VolumeType `json:"type"`
}

// Validate ensures that all required fields are populated.
func (conf *VolumeConfig) Validate() error {
	if conf.Source == "" {
		return fmt.Errorf("missing required field: volume.source")
	}

	if conf.Mount == "" {
		return fmt.Errorf("missing required field: volume.mount")
	}

	if conf.Type == "" {
		return fmt.Errorf("missing required field: volume.type")
	}

	return nil
}

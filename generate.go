package bob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/nickcorin/bob/templates"
)

//go:generate stringer -type FileType -trimprefix FileType

// FileType defines an enumerated constant of valid filetypes that bob can
// generate.
type FileType int

const (
	// FileTypeUnknown is an invalid filetype and usually indicates missing or
	// incorrect data.
	FileTypeUnknown FileType = iota

	// FileTypeDockerCompose defines a Docker Compose file.
	FileTypeDockerCompose

	// FileTypeDockerfile defines a Dockerfile.
	FileTypeDockerfile

	// FileTypeMakefile defines a Makefile.
	FileTypeMakefile

	// FileTypeServiceDiscovery defines a Prometheus Service Discovery.
	FileTypeServiceDiscovery

	// Must be last.
	fileTypeSentinel
)

func FileTypeFromString(s string) FileType {
	for ft := FileTypeUnknown + 1; ft.Valid(); ft++ {
		if strings.EqualFold(s, ft.String()) {
			return ft
		}
	}

	return FileTypeUnknown
}

// Valid returns whether ft is a declared FileType constant.
func (ft FileType) Valid() bool {
	return ft > FileTypeUnknown && ft < fileTypeSentinel
}

type generator struct {
	Filename string
	Fn       func(configs ...*Config) ([]byte, error)
	Type     FileType
}

// IsServiceFile returns whether the generated file is required for each
// service.
func (g generator) IsServiceFile() bool {
	return g.Type != FileTypeDockerCompose
}

var generators = map[FileType]generator{
	FileTypeDockerCompose: {
		Filename: "docker-compose.yml",
		Fn:       generateDockerCompose,
		Type:     FileTypeDockerCompose,
	},
	FileTypeDockerfile: {
		Filename: "Dockerfile",
		Fn:       generateDockerfile,
		Type:     FileTypeDockerfile,
	},
	FileTypeMakefile: {
		Filename: "Makefile",
		Fn:       generateMakefile,
		Type:     FileTypeMakefile,
	},
	FileTypeServiceDiscovery: {
		Filename: "%s.sd.json",
		Fn:       generateServiceDiscovery,
		Type:     FileTypeServiceDiscovery,
	},
}

func generateImageName(serviceName string) (string, error) {
	hash, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("get commit hash: %w", err)
	}

	image := fmt.Sprintf("%s:%s", serviceName, hash)

	// We must trim to remove the lurking '\n' characters.
	return strings.TrimSpace(image), nil
}

func findServiceConfigs(buildDir string) ([]string, error) {
	var configs []string
	if err := filepath.WalkDir(buildDir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("skipping directory due to error %s: %s", path, err)
				return fs.SkipDir
			}

			if d.IsDir() {
				// We want to enter all directories immediately.
				return nil
			}

			if strings.EqualFold(d.Name(), "bob.json") {
				// Keep track of all config files that we find.
				abs, err := filepath.Abs(path)
				if err != nil {
					return fmt.Errorf("absolute path: %w", err)
				}

				configs = append(configs, abs)
			}

			return nil
		}); err != nil {
		return nil, fmt.Errorf("walk build dir: %w", err)
	}

	return configs, nil
}

func GenerateService(configPath string) error {
	// Convert the path to an absolute path.
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("config absolute path: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("stat config: %w", err)
	}

	conf, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Iteratively generate service files.
	for _, requirement := range conf.Requires {
		ft := FileTypeFromString(requirement)
		if !ft.Valid() {
			log.Printf("skipping invalid requirement: %s", requirement)
		}

		g, ok := generators[ft]
		if !ok {
			return fmt.Errorf("missing generator for file type: %s", ft)
		}

		if !g.IsServiceFile() {
			// Skip non-service files.
			continue
		}

		data, err := g.Fn(conf)
		if err != nil {
			return fmt.Errorf("generate file: %w", err)
		}

		// We need to go up one directory in order to find the build directory.
		buildDir := filepath.Dir(filepath.Dir(configPath))

		var out string
		if ft == FileTypeServiceDiscovery {
			out = filepath.Join(buildDir, "prometheus",
				fmt.Sprintf(g.Filename, conf.Service.Name))
		} else {
			out = filepath.Join(buildDir, conf.Service.Name, g.Filename)
		}

		if _, err = write(data, out); err != nil {
			return fmt.Errorf("write file to disk: %w", err)
		}
	}

	return nil
}

func GenerateDockerCompose(buildDir string) error {
	configPaths, err := findServiceConfigs(buildDir)
	if err != nil {
		return fmt.Errorf("find service configs: %w", err)
	}

	var configs []*Config
	for _, path := range configPaths {
		conf, err := LoadConfig(path)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		configs = append(configs, conf)
	}

	g, ok := generators[FileTypeDockerCompose]
	if !ok {
		return fmt.Errorf("missing generator for docker compose")
	}

	data, err := g.Fn(configs...)
	if err != nil {
		return fmt.Errorf("generate docker-compose: %w", err)
	}

	if _, err = write(data, filepath.Join(buildDir, g.Filename)); err != nil {
		return fmt.Errorf("write file to disk: %w", err)
	}

	return nil
}

func generateDockerCompose(configs ...*Config) ([]byte, error) {
	t, err := template.New("docker-compose").Parse(templates.Compose)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	params := struct {
		Configs      []Config
		NamedVolumes []VolumeConfig
	}{Configs: make([]Config, 0)}

	for _, c := range configs {
		params.Configs = append(params.Configs, *c)

		// Pre-process the configs to find all the named volumes for easier
		// generation in the template.
		for _, v := range c.Docker.Volumes {
			if v.Type == VolumeTypeBind {
				continue
			}

			if params.NamedVolumes == nil {
				params.NamedVolumes = make([]VolumeConfig, 0)
			}

			params.NamedVolumes = append(params.NamedVolumes, v)
		}
	}

	var data bytes.Buffer
	if err = t.Execute(&data, params); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	// Replace all tab characters with 4 spaces. Tab characters are not allowed
	// in docker-compose.yml files.
	clean := strings.ReplaceAll(data.String(), "\t", "    ")

	return []byte(clean), nil
}

func generateDockerfile(configs ...*Config) ([]byte, error) {
	conf := configs[0]

	t, err := template.New(conf.Service.Name).Parse(templates.Dockerfile)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var data bytes.Buffer
	if err = t.Execute(&data, conf); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return data.Bytes(), nil
}

func generateMakefile(configs ...*Config) ([]byte, error) {
	conf := configs[0]

	t, err := template.New(conf.Service.Name).Parse(templates.Makefile)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var data bytes.Buffer
	if err = t.Execute(&data, conf); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return data.Bytes(), nil
}

func generateServiceDiscovery(configs ...*Config) ([]byte, error) {
	conf := configs[0]

	if conf.Service.MetricsPort == 0 {
		return nil, fmt.Errorf("metrics port required for service discovery")
	}

	params := []struct {
		Targets []string          `json:"targets"`
		Labels  map[string]string `json:"labels"`
	}{{
		[]string{
			fmt.Sprintf("%s:%d", conf.Service.Host, conf.Service.MetricsPort),
		},
		map[string]string{"service": conf.Service.Name},
	}}

	data, err := json.MarshalIndent(params, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("encode static config: %w", err)
	}

	return data, nil
}

func write(data []byte, outFile string) ([]byte, error) {
	f, err := os.Create(outFile)
	if err != nil {
		return nil, fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return nil, fmt.Errorf("write to output file: %w", err)
	}

	fmt.Printf("wrote %s successfully.\n", outFile)

	return data, nil
}

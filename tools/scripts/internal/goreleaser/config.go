package goreleaser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version          int              `yaml:"version"`
	ProjectName      string           `yaml:"project_name"`
	Builds           []Build          `yaml:"builds"`
	Archives         []Archive        `yaml:"archives"`
	UniversalBinaries []UniversalBinary `yaml:"universal_binaries"`
	Checksum         Checksum         `yaml:"checksum"`
	Release          Release          `yaml:"release"`
	Changelog        Changelog        `yaml:"changelog"`
	Signs            []Sign            `yaml:"signs"`
}

type Build struct {
	ID      string   `yaml:"id"`
	Main    string   `yaml:"main"`
	Binary  string   `yaml:"binary"`
	Goos    []string `yaml:"goos"`
	Goarch  []string `yaml:"goarch"`
	Tool    string   `yaml:"tool,omitempty"`
	Goarm   []string `yaml:"goarm,omitempty"`
	Goarm64 []string `yaml:"goarm64,omitempty"`
}

type Archive struct {
	ID           string   `yaml:"id"`
	IDs          []string `yaml:"ids"`
	Formats      []string `yaml:"formats"`
	NameTemplate string   `yaml:"name_template"`
	Files        []string `yaml:"files"`
}

type UniversalBinary struct {
	ID           string   `yaml:"id"`
	IDs          []string `yaml:"ids"`
	NameTemplate string   `yaml:"name_template"`
	Replace      bool     `yaml:"replace"`
}

type Checksum struct {
	NameTemplate string `yaml:"name_template"`
}

type Release struct {
	Draft                 bool   `yaml:"draft"`
	ReplaceExistingDraft  bool   `yaml:"replace_existing_draft"`
	Prerelease             string `yaml:"prerelease"`
}

type Changelog struct {
	Sort    string  `yaml:"sort"`
	Filters Filters `yaml:"filters"`
}

type Filters struct {
	Exclude []string `yaml:"exclude"`
}

type Sign struct {
	Artifacts string   `yaml:"artifacts"`
	Cmd       string   `yaml:"cmd"`
	Args      []string `yaml:"args"`
}

func Generate(legacyRoot string) error {
	legacyGo := filepath.ToSlash(filepath.Join(legacyRoot, "bin", goExecutable()))
	config := newConfig(legacyGo)

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(filepath.Join("..", "..", ".goreleaser.yaml"), data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func goExecutable() string {
	if runtime.GOOS == "windows" {
		return "go.exe"
	}
	return "go"
}

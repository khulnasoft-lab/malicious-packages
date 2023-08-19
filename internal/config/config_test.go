package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/khulnasoft-lab/infected-packages/internal/config"
	"github.com/khulnasoft-lab/infected-packages/internal/source"
)

const validConfigYAML = `
id-prefix: TEST
infected-path: "mal/"
false-positive-path: "false-positives/"
sources:
- id: all
  bucket: file://test-bucket/
  prefix: infected
  lookback-entries: 123
- id: default
`

func TestReadYAML(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() = %v; want no error", err)
	}
	c, err := config.ReadYAML(bytes.NewBufferString(validConfigYAML))
	if err != nil {
		t.Fatalf("ReadYAML = %v; want no error", err)
	}
	if got, want := len(c.Sources), 2; got != want {
		t.Errorf("len(Sources) = %v; want %v", got, want)
	}
	if got, want := c.IDPrefix, "TEST"; got != want {
		t.Errorf("IDPrefix = %v; want %v", got, want)
	}
	if got, want := c.InfectedPath, filepath.Join(wd, "mal"); got != want {
		t.Errorf("InfectedPath = %v; want %v", got, want)
	}
	if got, want := c.FalsePositivePath, filepath.Join(wd, "false-positives"); got != want {
		t.Errorf("FalsePositivePath = %v; want %v", got, want)
	}
}

func TestReadYAML_Error(t *testing.T) {
	_, err := config.ReadYAML(bytes.NewBufferString(""))
	if err == nil {
		t.Fatal("ReadYAML = nil; want an error", err)
	}
}

func TestReadYAML_Invalid(t *testing.T) {
	_, err := config.ReadYAML(bytes.NewBufferString("sources: hello"))
	if err == nil {
		t.Fatal("ReadYAML = nil; want an error", err)
	}
}

func TestPaths(t *testing.T) {
	got := getTestConfig().Paths()
	want := []string{"./false/positives/", "./mal/path/"}
	sort.StringSlice(got).Sort() // Sort to eliminate any ordering issues.
	if !slices.Equal(got, want) {
		t.Errorf("c.Paths() = %v; want %v", got, want)
	}
}

func TestInit(t *testing.T) {
	if err := getTestConfig().Init(); err != nil {
		t.Errorf("Init() = %v; want no error", err)
	}
}

func TestInit_NoIDPrefix(t *testing.T) {
	c := getTestConfig()
	c.IDPrefix = ""
	if err := c.Init(); err == nil {
		t.Error("Init() = nil; want an error")
	}
}

func TestInit_NoInfectedPath(t *testing.T) {
	c := getTestConfig()
	c.InfectedPath = ""
	if err := c.Init(); err == nil {
		t.Error("Init() = nil; want an error")
	}
}

func TestInit_NoFalsePositivePath(t *testing.T) {
	c := getTestConfig()
	c.FalsePositivePath = ""
	if err := c.Init(); err == nil {
		t.Error("Init() = nil; want an error")
	}
}

func getTestConfig() *config.Config {
	return &config.Config{
		IDPrefix:          "FOO",
		InfectedPath:      "./mal/path/",
		FalsePositivePath: "./false/positives/",
		Sources: []*source.Source{
			{
				ID: "source1",
			},
			{
				ID: "source2",
			},
		},
	}
}

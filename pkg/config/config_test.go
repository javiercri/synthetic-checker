package config

import (
	"os"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestTemplatedConfig(t *testing.T) {
	yamlCfg := `httpChecks:
  test:
    url: https://{{ .Env.FAKE_HOST }}/foo
    headers:
      Authorization: Bearer {{ .Env.FAKE_TOKEN }}
    interval: 10s`

	os.Setenv("FAKE_HOST", "fake.com")
	os.Setenv("FAKE_TOKEN", "T0p$€cRet")
	var cfg Config
	err := yaml.Unmarshal([]byte(yamlCfg), &cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HTTPChecks["test"].URL != "https://fake.com/foo" {
		t.Errorf("unexpected URL: wanted: %q; got: %q", cfg.HTTPChecks["test"].URL, "https://fake.com/foo")
	}

	if cfg.HTTPChecks["test"].Headers["Authorization"] != "Bearer T0p$€cRet" {
		t.Error("unexpected Authorization header")
	}
}

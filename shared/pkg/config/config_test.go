package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	AppName string        `yaml:"app_name" env:"APP_NAME" env-default:"my_app"`
	Port    int           `yaml:"port" env:"PORT" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"5s"`
}

func TestMustLoad_Success(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		yamlContent string
		expectedCfg TestConfig
	}{
		{
			name: "Успешная загрузка только из ENV (файл не указан)",
			envVars: map[string]string{
				"APP_NAME": "env_app",
				"PORT":     "8080",
				"TIMEOUT":  "10s",
			},
			yamlContent: "",
			expectedCfg: TestConfig{
				AppName: "env_app",
				Port:    8080,
				Timeout: 10 * time.Second,
			},
		},
		{
			name:    "Успешная загрузка из YAML (ENV пустые)",
			envVars: map[string]string{},
			yamlContent: `
app_name: "yaml_app"
port: 9090
timeout: "15s"
`,
			expectedCfg: TestConfig{
				AppName: "yaml_app",
				Port:    9090,
				Timeout: 15 * time.Second,
			},
		},
		{
			name: "ENV перекрывает значения из YAML",
			envVars: map[string]string{
				"PORT": "8080",
			},
			yamlContent: `
app_name: "yaml_app"
port: 9090
timeout: "15s"
`,
			expectedCfg: TestConfig{
				AppName: "yaml_app",
				Port:    8080,
				Timeout: 15 * time.Second,
			},
		},
		{
			name: "Использование дефолтных значений",
			envVars: map[string]string{
				"PORT": "3000",
			},
			yamlContent: "",
			expectedCfg: TestConfig{
				AppName: "my_app",
				Port:    3000,
				Timeout: 5 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			if tt.yamlContent != "" {
				tempDir := t.TempDir()
				tempFile := filepath.Join(tempDir, "config.yaml")

				err := os.WriteFile(tempFile, []byte(tt.yamlContent), 0644)
				require.NoError(t, err)

				t.Setenv("CONFIG_PATH", tempFile)
			} else {
				t.Setenv("CONFIG_PATH", "")
			}

			var cfg TestConfig
			MustLoad(&cfg)

			assert.Equal(t, tt.expectedCfg, cfg)
		})
	}
}

// TestMustLoad_FatalCases проверяет, что при неверной конфигурации
// вызывается log.Fatalf (os.Exit(1)).
func TestMustLoad_FatalCases(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		var cfg TestConfig
		MustLoad(&cfg) // Должен сделать log.Fatalf
		return
	}

	tests := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "Ошибка: файл конфига по пути CONFIG_PATH не существует",
			envVars: map[string]string{
				"CONFIG_PATH": "/path/to/invalid/or/non/existent/config.yaml",
				"PORT":        "8080",
			},
		},
		{
			name: "Ошибка: не задано обязательное поле env-required",
			envVars: map[string]string{
				"CONFIG_PATH": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestMustLoad_FatalCases$")

			cmd.Env = append(os.Environ(), "BE_CRASHER=1")
			for k, v := range tt.envVars {
				cmd.Env = append(cmd.Env, k+"="+v)
			}

			err := cmd.Run()

			if e, ok := err.(*exec.ExitError); ok && !e.Success() {
				return
			}

			t.Fatalf("Ожидалось падение через log.Fatalf, но процесс завершился без ошибок: %v", err)
		})
	}
}

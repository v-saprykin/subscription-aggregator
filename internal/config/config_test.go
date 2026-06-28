package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	setEmptyConfigEnv(t)

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := Config{
		AppEnv:      "local",
		HTTPAddr:    ":8080",
		LogLevel:    "info",
		DatabaseURL: "",
	}
	if got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}

func TestLoadEnvironmentOverrides(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  Config
	}{
		{
			name:  "app environment",
			key:   "APP_ENV",
			value: "test",
			want: Config{
				AppEnv:      "test",
				HTTPAddr:    ":8080",
				LogLevel:    "info",
				DatabaseURL: "",
			},
		},
		{
			name:  "HTTP address",
			key:   "HTTP_ADDR",
			value: "127.0.0.1:9090",
			want: Config{
				AppEnv:      "local",
				HTTPAddr:    "127.0.0.1:9090",
				LogLevel:    "info",
				DatabaseURL: "",
			},
		},
		{
			name:  "log level",
			key:   "LOG_LEVEL",
			value: "debug",
			want: Config{
				AppEnv:      "local",
				HTTPAddr:    ":8080",
				LogLevel:    "debug",
				DatabaseURL: "",
			},
		},
		{
			name:  "database URL",
			key:   "DATABASE_URL",
			value: "postgres://app:secret@localhost:5432/subscriptions",
			want: Config{
				AppEnv:      "local",
				HTTPAddr:    ":8080",
				LogLevel:    "info",
				DatabaseURL: "postgres://app:secret@localhost:5432/subscriptions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEmptyConfigEnv(t)
			t.Setenv(tt.key, tt.value)

			got, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func setEmptyConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{"APP_ENV", "HTTP_ADDR", "LOG_LEVEL", "DATABASE_URL"} {
		t.Setenv(key, "")
	}
}

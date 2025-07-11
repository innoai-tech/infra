package internal

import (
	"strings"
	"unicode"
)

func EnvVarsFromEnviron(environ []string) EnvVars {
	envVars := EnvVars{}
	for _, kv := range environ {
		parts := strings.SplitN(kv, "=", 2)
		envVars.Add(parts[0], parts[1])
	}
	return envVars
}

type EnvVars map[string]string

func (envVars EnvVars) Add(key, value string) {
	envVars[toUpperDigit(key)] = value
}

func (envVars EnvVars) Get(envVar string) (string, bool) {
	v, ok := envVars[toUpperDigit(envVar)]
	return v, ok
}

func toUpperDigit(s string) string {
	b := make([]byte, 0, len(s))
	for _, c := range s {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			b = append(b, byte(unicode.ToUpper(c)))
		}
	}
	return string(b)
}

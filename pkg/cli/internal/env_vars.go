package internal

import (
	"strings"
	"unicode"
)

// EnvVarsFromEnviron 从环境变量键值对列表中构建 EnvVars。
func EnvVarsFromEnviron(environ []string) EnvVars {
	envVars := EnvVars{}
	for _, kv := range environ {
		parts := strings.SplitN(kv, "=", 2)
		envVars.Add(parts[0], parts[1])
	}
	return envVars
}

// EnvVars 表示环境变量的键值映射。
type EnvVars map[string]string

// Add 添加或覆盖一个环境变量。
func (envVars EnvVars) Add(key, value string) {
	envVars[toUpperDigit(key)] = value
}

// Get 获取指定环境变量的值。
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

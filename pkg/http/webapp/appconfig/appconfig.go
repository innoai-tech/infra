package appconfig

import (
	"sort"
	"strings"
)

// ParseAppConfig 从字符串解析应用配置键值对。
func ParseAppConfig(s string) AppConfig {
	parts := strings.Split(s, ",")

	c := AppConfig{}

	for i := range parts {
		kv := strings.Split(parts[i], "=")

		if kv[0] == "" {
			continue
		}

		if len(kv) == 2 {
			c[kv[0]] = kv[1]
		} else {
			c[kv[0]] = ""
		}
	}

	return c
}

// AppConfig 表示前端应用运行时的键值配置。
type AppConfig map[string]string

// EnvVarPrefix 是环境变量注入 AppConfig 时使用的前缀。
const EnvVarPrefix = "APP_CONFIG__"

// LoadFromEnviron 从环境变量键值对中加载配置。
func (c AppConfig) LoadFromEnviron(kv []string) {
	for i := range kv {
		keyValue := strings.SplitN(kv[i], "=", 2)
		key := keyValue[0]
		if len(keyValue) >= 2 && strings.HasPrefix(key, EnvVarPrefix) {
			c[key[len(EnvVarPrefix):]] = keyValue[1]
		}
	}
}

// String 返回按键排序后的配置字符串表示。
func (c AppConfig) String() string {
	keys := make([]string, 0)

	for k := range c {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	b := strings.Builder{}

	for i, k := range keys {
		if i != 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(c[k])
	}

	return b.String()
}

package internal

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestEnvVars(t *testing.T) {
	t.Run("GIVEN environ with ENV_VAR_1", func(t *testing.T) {
		envVars := EnvVarsFromEnviron([]string{
			"env_var_1=1",
		})

		t.Run("WHEN get by full name with underscores", func(t *testing.T) {
			v, _ := envVars.Get("ENV_VAR_1")

			Then(t, "it should return the value",
				Expect(v, Equal("1")),
			)
		})

		t.Run("WHEN get by name without underscores", func(t *testing.T) {
			v, _ := envVars.Get("ENVVAR1")

			Then(t, "it should still return the normalized value",
				Expect(v, Equal("1")),
			)
		})
	})
}

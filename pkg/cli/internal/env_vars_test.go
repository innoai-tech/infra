package internal

import (
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

func TestEnvVars(t *testing.T) {
	bdd.FromT(t).Given("environ", func(b bdd.T) {
		envVars := EnvVarsFromEnviron([]string{
			"env_var_1=1",
		})

		b.When("get by full env var", func(b bdd.T) {
			v, _ := envVars.Get("ENV_VAR_1")
			b.Then("success",
				bdd.Equal("1", v),
			)
		})

		b.When("get by full env var without lodash", func(b bdd.T) {
			v, _ := envVars.Get("ENVVAR1")
			b.Then("success",
				bdd.Equal("1", v),
			)
		})
	})
}

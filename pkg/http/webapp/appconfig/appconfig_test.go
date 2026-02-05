package appconfig

import (
	"testing"

	testingv2 "github.com/octohelm/x/testing/v2"
)

func TestAppConfig(t *testing.T) {
	t.Run("GIVEN app config", func(t *testing.T) {
		ac := AppConfig{
			"KEY1": "VALUE1",
			"KEY2": "VALUE2",
		}

		testingv2.Then(t, "stringify as expect",
			testingv2.Expect(
				ac.String(),
				testingv2.Equal("KEY1=VALUE1,KEY2=VALUE2"),
			),
		)

		t.Run("WHEN parse", func(t *testing.T) {
			testingv2.Then(t, "parsed should be same as given",
				testingv2.Expect(
					ParseAppConfig(ac.String()),
					testingv2.Equal(ac),
				),
			)
		})

		t.Run("WHEN load from environ", func(t *testing.T) {
			e := AppConfig{}
			e.LoadFromEnviron([]string{
				"APP_CONFIG__KEY1=VALUE1",
				"APP_CONFIG__KEY2=VALUE2",
				"XX=Value=1", // should ignore
			})

			testingv2.Then(t, "loaded should be same as given",
				testingv2.Expect(e,
					testingv2.Equal(ac),
				),
			)
		})
	})
}

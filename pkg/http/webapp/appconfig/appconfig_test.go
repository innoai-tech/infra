package appconfig

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestAppConfig(t *testing.T) {
	ac := AppConfig{
		"KEY1": "VALUE1",
		"KEY2": "VALUE2",
	}

	t.Run("#String", func(t *testing.T) {
		testingx.Expect(t, ac.String(), testingx.Equal("KEY1=VALUE1,KEY2=VALUE2"))
	})

	t.Run("#LoadFromEnviron", func(t *testing.T) {
		e := AppConfig{}
		e.LoadFromEnviron([]string{"APP_CONFIG__KEY1=VALUE1", "APP_CONFIG__KEY2=VALUE2", "XX=Value=1"})

		testingx.Expect(t, e, testingx.Equal(ac))
	})

	t.Run("ParseAppConfig", func(t *testing.T) {
		e := ParseAppConfig(ac.String())
		testingx.Expect(t, e, testingx.Equal(ac))
	})
}

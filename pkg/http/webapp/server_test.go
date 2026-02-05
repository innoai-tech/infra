package webapp

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestOpt(t *testing.T) {
	t.Run("with /", func(t *testing.T) {
		o := (&opt{}).build(WithBaseHref("/"))

		t.Run("without header X-App-Base-Href", func(t *testing.T) {
			Then(t, "base href keep static",
				Expect(
					o.resolveBaseHref(""),
					Equal("/"),
				),
			)
		})

		t.Run("with header X-App-Base-Href", func(t *testing.T) {
			Then(t, "base href use header value",
				Expect(
					o.resolveBaseHref("/clusters/test/x-app/"),
					Equal("/clusters/test/x-app/"),
				),
			)
		})
	})

	t.Run("with /base/", func(t *testing.T) {
		o := (&opt{}).build(WithBaseHref("/base/"))

		t.Run("without header X-App-Base-Href", func(t *testing.T) {
			Then(t, "base href keep static",
				Expect(
					o.resolveBaseHref(""),
					Equal("/base/"),
				),
			)
		})

		t.Run("with header X-App-Base-Href", func(t *testing.T) {
			Then(t, "base href use header value",
				Expect(
					o.resolveBaseHref("/clusters/test/x-app/"),
					Equal("/clusters/test/x-app/base/"),
				),
			)
		})
	})
}

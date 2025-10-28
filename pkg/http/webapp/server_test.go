package webapp

import (
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

func TestOpt(t *testing.T) {
	b := bdd.FromT(t)

	b.Given("with /", func(b bdd.T) {
		o := (&opt{}).build(WithBaseHref("/"))

		b.When("without header X-App-Base-Href", func(b bdd.T) {
			b.Then("keep static",
				bdd.Equal("/", o.resolveBaseHref("")),
			)
		})

		b.When("with header X-App-Base-Href", func(b bdd.T) {
			b.Then("use header value",
				bdd.Equal("/clusters/test/x-app/", o.resolveBaseHref("/clusters/test/x-app/")),
			)
		})
	})

	b.Given("with /base/", func(b bdd.T) {
		o := (&opt{}).build(WithBaseHref("/base/"))

		b.When("without header X-App-Base-Href", func(b bdd.T) {
			b.Then("keep static",
				bdd.Equal("/base/", o.resolveBaseHref("")),
			)
		})

		b.When("with header X-App-Base-Href", func(b bdd.T) {
			b.Then("use header value",
				bdd.Equal("/clusters/test/x-app/base/", o.resolveBaseHref("/clusters/test/x-app/")),
			)
		})
	})
}

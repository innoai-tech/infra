package cli

import (
	"bytes"
	"context"
	"fmt"
	"github.com/octohelm/x/slices"
	"os"
	"path"
	"strings"

	cueformat "cuelang.org/go/cue/format"
	"github.com/octohelm/gengo/pkg/camelcase"
	"github.com/octohelm/gengo/pkg/gengo"
)

func (c *C) dumpK8sConfiguration(ctx context.Context, dest string) error {
	if c.i.Component == nil {
		return nil
	}

	pkgName := strings.ToLower(camelcase.LowerCamelCase(c.i.App.Name))
	componentName := camelcase.LowerKebabCase(c.i.Component.Name)

	dest = path.Join(dest, pkgName)

	kind := "Deployment"

	if k := c.i.Component.Options.Get("kind"); k != "" {
		kind = k
	}

	b := bytes.NewBuffer(nil)

	_, _ = fmt.Fprintf(b, `
package %s

import (
	kubepkg "github.com/octohelm/kubepkgspec/cuepkg/kubepkg"
)

#%s: kubepkg.#KubePkg & {
metadata: {
	name: string | *%q
}
spec: {
	version: _

	deploy: kind: %q
`,
		pkgName,
		gengo.UpperCamelCase(componentName),
		componentName,
		kind,
	)

	if kind == "Deployment" {
		_, _ = fmt.Fprintf(b, `
	deploy: spec: replicas: _ | *1
`)
	}

	toComment := func(s string) string {
		return strings.Join(
			slices.Map(strings.Split(s, "\n"), func(line string) string {
				return "// " + line
			}),
			"\n",
		)
	}

	var flagExposes []*flagVar

	i := 0
	for _, f := range c.flagVars {
		if f.Expose != "" {
			flagExposes = append(flagExposes, f)
			continue
		}

		if i == 0 {
			b.WriteByte('\n')
		}

		if f.Required {
			_, _ = fmt.Fprintf(b,
				`
%s
config: %q: string
`, toComment(f.Desc), f.EnvVar)
			continue
		}

		_, _ = fmt.Fprintf(b,
			`
%s 
config: %q: string | *%q
`, toComment(f.Desc), f.EnvVar, f.string())

		i++
	}

	if len(flagExposes) > 0 {
		_, _ = fmt.Fprintf(b, `
	services: "#": {
		ports: containers.%q.ports
	}
`, componentName)

		_, _ = fmt.Fprintf(b, `
	containers: %q: {
`, componentName)

		for i := range flagExposes {
			portName := "http"
			if i != 0 {
				portName = gengo.LowerKebabCase(flagExposes[i].Name)
			}

			parts := strings.Split(flagExposes[i].String(), ":")

			_, _ = fmt.Fprintf(b, `
		ports: %q: _ | *%v
`, portName, parts[1])
			_, _ = fmt.Fprintf(b, `
		env: %q: _ | *":\(ports.%q)"
`, flagExposes[i].EnvVar, portName)

			if i == 0 {
				// only use first as probe
				_, _ = fmt.Fprintf(b, `
		readinessProbe: {
			httpGet: {
				path: "/", 
				port: ports.%q
				scheme: "HTTP"
			}
			initialDelaySeconds: _ | *5
            timeoutSeconds:      _ | *1
            periodSeconds:       _ | *10
            successThreshold:    _ | *1
            failureThreshold:    _ | *3
		}
		livenessProbe: readinessProbe
`, portName)
			}
		}
		_, _ = fmt.Fprintf(b, `  
	}
`)
	}

	_, _ = fmt.Fprintf(b, `  
	containers: %q: {
		image: {
			name: _ | *"%v/%v"
			tag:  _ | *"\(version)"
		}
`, componentName, c.i.App.ImageNamespace, c.i.App.Name)

	_, _ = fmt.Fprintf(b, `
		args: [
`)
	for _, n := range c.cmdPath {
		_, _ = fmt.Fprintf(b, `		%q,`, n)
	}
	_, _ = fmt.Fprintf(b, `
		]
	}
}
}`)

	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}

	data, err := cueformat.Source(b.Bytes(), cueformat.Simplify())
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(dest, fmt.Sprintf("%s.cue", componentName)), data, 0600)
}

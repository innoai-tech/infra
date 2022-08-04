package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
	"github.com/octohelm/gengo/pkg/gengo"
)

func (c *C) dumpK8sConfiguration(ctx context.Context, dest string) error {
	if c.i.Component == "" {
		return nil
	}

	pkgName := strings.ToLower(camelcase.LowerCamelCase(c.i.App.Name))
	componentName := camelcase.LowerKebabCase(c.i.Component)

	dest = path.Join(dest, pkgName)

	b := bytes.NewBuffer(nil)
	_, _ = fmt.Fprintf(b, `
package %s

import (
	"github.com/innoai-tech/runtime/cuepkg/kube"
)

#%s: kube.#App & {
	app: {
		name: string | *%q
	}
`,
		pkgName,
		gengo.UpperCamelCase(componentName),
		componentName,
	)

	var flagExposes []*flagVar

	for i := range c.flagVars {
		f := c.flagVars[i]

		if f.Expose != "" {
			flagExposes = append(flagExposes, f)
			continue
		}

		if f.Required {
			_, _ = fmt.Fprintf(b, `  
	config: %q: string
`, f.EnvVar)
			continue
		}

		_, _ = fmt.Fprintf(b, `  
	config: %q: string | *%q
`, f.EnvVar, f.string())
	}

	if len(flagExposes) > 0 {
		_, _ = fmt.Fprintf(b, `
	services: "\(app.name)": {
		selector: "app": app.name
		ports:     containers.%q.ports
	}
`, componentName)

		_, _ = fmt.Fprintf(b, `
	containers: %q: {
`, componentName)

		for i := range flagExposes {
			portName := "http"
			if i != 0 {
				portName = gengo.LowerKebabCase("http-" + flagExposes[i].Name)
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
		readinessProbe: kube.#ProbeHttpGet & {
			httpGet: {path: "/", port: ports.%q}
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
			tag:  _ | *"\(app.version)"
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
}`)

	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(path.Join(dest, fmt.Sprintf("%s.cue", componentName)), b.Bytes(), os.ModePerm)
}

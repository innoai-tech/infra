package cli

import (
	"bytes"
	"context"
	"fmt"
	"github.com/octohelm/gengo/pkg/camelcase"
	"github.com/octohelm/gengo/pkg/gengo"
	"os"
	"path"
	"strings"
)

func (c *C) dumpK8sConfiguration(ctx context.Context, appName string, dest string) error {
	names := make([]string, 0)

	cmd := c

	for cmd.parent != nil {
		names = append([]string{cmd.Name}, names...)

		cmd = cmd.parent.CmdInfo()
	}

	subAppName := camelcase.LowerKebabCase(strings.Join(append([]string{appName}, names...), "-"))
	pkgName := strings.ToLower(camelcase.LowerCamelCase(subAppName))
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
		gengo.UpperCamelCase(subAppName),
		subAppName,
	)

	for i := range c.flagVars {
		f := c.flagVars[i]

		if f.Required {
			_, _ = fmt.Fprintf(b, `  config: %q: %s
`, f.EnvVar, f.Type())
			continue
		}

		if f.Type() == "string" {
			_, _ = fmt.Fprintf(b, `  config: %q: %v | *%q
`, f.EnvVar, f.Type(), f.Value.Interface())
			continue
		}

		_, _ = fmt.Fprintf(b, `  config: %q: %v | *%v
`, f.EnvVar, f.Type(), f.Value.Interface())
	}

	_, _ = fmt.Fprintf(b, `
	services: "\(app.name)": {
		selector: "app": app.name
		ports:     containers.%q.ports
	}

	containers: %q: {
`, subAppName, subAppName)

	_, _ = fmt.Fprintf(b, `  
		image: {
			name: _ | *"ghcr.io/octohelm/%v"
			tag:  _ | *"\(app.version)"
		}
		ports: http: _ | *80
		readinessProbe: kube.#ProbeHttpGet & {
			httpGet: {path: "/", port: ports.http}
		}
		livenessProbe: readinessProbe
		args: [
`, subAppName)
	for _, n := range names {
		_, _ = fmt.Fprintf(b, `		%q,`, n)
	}
	_, _ = fmt.Fprintf(b, `
  		]
	}
}
`)

	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(path.Join(dest, "app.cue"), b.Bytes(), os.ModePerm)
}

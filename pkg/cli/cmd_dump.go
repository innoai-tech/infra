package cli

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/x/slices"

	"github.com/innoai-tech/infra/pkg/cli/internal"
)

// cueBuilder 辅助生成带一致缩进的 CUE 代码，统一使用 tab。
type cueBuilder struct {
	b     strings.Builder
	depth int
}

func (w *cueBuilder) indent() {
	for i := 0; i < w.depth; i++ {
		w.b.WriteByte('\t')
	}
}

func (w *cueBuilder) line(format string, args ...any) {
	w.indent()
	fmt.Fprintf(&w.b, format, args...)
	w.b.WriteByte('\n')
}

func (w *cueBuilder) block(format string, args ...any) {
	w.indent()
	fmt.Fprintf(&w.b, format, args...)
	w.b.WriteString(" {\n")
	w.depth++
}

func (w *cueBuilder) end() {
	w.depth--
	w.indent()
	w.b.WriteString("}\n")
}

func (c *C) dumpK8sConfiguration(ctx context.Context, dest string) error {
	if c.info.Component == nil {
		return nil
	}

	pkgName := strings.ToLower(camelcase.LowerCamelCase(c.info.App.Name))
	componentName := camelcase.LowerKebabCase(c.info.Component.Name)

	dest = path.Join(dest, pkgName)

	kind := "Deployment"

	if k := c.info.Component.Options.Get("kind"); k != "" {
		kind = k
	}

	w := &cueBuilder{}

	w.line("package %s", pkgName)
	w.b.WriteByte('\n')
	w.line("import (")
	w.depth++
	w.line("kubepkg %q", "github.com/octohelm/kubepkgspec/cuepkg/kubepkg")
	w.depth--
	w.line(")")
	w.b.WriteByte('\n')

	w.block("#%s: kubepkg.#KubePkg &", gengo.UpperCamelCase(componentName))

	w.block("metadata:")
	w.line("name: string | *%q", componentName)
	w.end()

	w.block("spec:")
	w.line("version: _")
	w.b.WriteByte('\n')

	w.line("deploy: kind: %q", kind)

	if kind == "Deployment" {
		w.line("deploy: spec: replicas: _ | *1")
		w.b.WriteByte('\n')
	}

	toComment := func(s string) string {
		return strings.Join(
			slices.Map(strings.Split(s, "\n"), func(line string) string {
				return "// " + line
			}),
			"\n",
		)
	}

	var flagExposes []*internal.FlagVar

	for _, f := range c.flagVars {
		if f.Expose != "" {
			flagExposes = append(flagExposes, f)
			continue
		}

		if f.Required {
			w.line("%s", toComment(f.Desc))
			w.line("config: %q: string", f.EnvVar)
			continue
		}

		w.line("%s", toComment(f.Desc))
		w.line("config: %q: string | *%q", f.EnvVar, f.DefaultValue())
	}

	if len(flagExposes) > 0 {
		w.b.WriteByte('\n')
		w.block("services: %q:", "#")
		w.line("ports: containers.%q.ports", componentName)
		w.end()

		w.b.WriteByte('\n')
		w.block("containers: %q:", componentName)

		for i := 0; i < len(flagExposes); i++ {
			portName := "http"
			if i != 0 {
				portName = gengo.LowerKebabCase(flagExposes[i].Name)
			}

			parts := strings.Split(flagExposes[i].String(), ":")

			w.line("ports: %q: _ | *%v", portName, parts[1])
			w.line("env: %q: _ | *\":\\(ports.%q)\"", flagExposes[i].EnvVar, portName)

			if i == 0 {
				// 仅第一个暴露端口用作探针
				w.block("readinessProbe:")

				w.block("httpGet:")
				w.line("path: _ | *%q", "/")
				w.line("port: _ | *ports.%q", portName)
				w.line("scheme: _ | *%q", "HTTP")
				w.end()

				w.line("initialDelaySeconds: _ | *5")
				w.line("timeoutSeconds:      _ | *1")
				w.line("periodSeconds:       _ | *10")
				w.line("successThreshold:    _ | *1")
				w.line("failureThreshold:    _ | *3")
				w.end()

				w.line("livenessProbe: readinessProbe")
			}
		}

		w.end()
		w.b.WriteByte('\n')
	}

	w.block("containers: %q:", componentName)
	w.block("image:")
	w.line("name: _ | *%q", fmt.Sprintf("%s/%s", c.info.App.ImageNamespace, c.info.App.Name))
	w.line("tag:  _ | *\"\\(version)\"")
	w.end()

	w.line("args: [")
	w.depth++
	for _, n := range c.cmdPath {
		w.line("%q,", n)
	}
	w.depth--
	w.line("]")
	w.end() // containers 块结束

	w.end() // spec 块结束
	w.end() // #Component 块结束

	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(path.Join(dest, fmt.Sprintf("%s.cue", componentName)), []byte(w.b.String()), 0o600)
}

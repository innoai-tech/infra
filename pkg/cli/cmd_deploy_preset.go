package cli

import (
	"context"
	"fmt"
	"go/format"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"

	"github.com/innoai-tech/infra/pkg/cli/internal"
)

// dumpDeployPreset 导出组件的 deploy.Container 预设 Go 代码。
func (c *C) dumpDeployPreset(ctx context.Context, dest string) error {
	if c.info.Component == nil {
		return nil
	}

	componentName := camelcase.LowerSnakeCase(c.info.Component.Name)
	dest = path.Join(dest, componentName)

	toComment := func(s string) string {
		if s == "" {
			return ""
		}
		return "// " + strings.ReplaceAll(s, "\n", "\n// ")
	}

	w := &gofmtBuilder{}

	w.line("package %s", componentName)
	w.b.WriteByte('\n')
	w.line("import %q", "github.com/innoai-tech/infra/pkg/deploy")
	w.b.WriteByte('\n')
	w.line("// Preset 返回预设的容器部署规格。")
	w.line("func Preset() *deploy.Container {")

	w.depth++

	w.line("c := &deploy.Container{")
	w.depth++

	imageName := fmt.Sprintf("%s/%s", c.info.App.ImageNamespace, c.info.App.Name)
	w.line("ImageName: %q,", imageName)
	w.line("Version: %q,", c.info.App.Version)

	if appName := c.info.App.Name; appName != "" {
		w.line("Command: []string{%q},", appName)
	}

	if len(c.cmdPath) > 0 {
		w.line("Args: []string{")
		w.depth++
		for _, n := range c.cmdPath {
			w.line("%q,", n)
		}
		w.depth--
		w.line("},")
	}

	var flagExposes []*internal.FlagVar
	var flagEnvs []*internal.FlagVar

	for _, f := range c.flagVars {
		if f.Expose != "" {
			flagExposes = append(flagExposes, f)
			continue
		}
		flagEnvs = append(flagEnvs, f)
	}

	if len(flagExposes) > 0 {
		w.line("Ports: map[string]deploy.Port{")
		w.depth++

		for _, f := range flagExposes {
			portName := f.Expose
			parts := strings.Split(f.String(), ":")
			port := parts[len(parts)-1]

			if f.Desc != "" {
				for line := range strings.Lines(toComment(f.Desc)) {
					w.line("%s", line)
				}
			}

			w.line("%q: {", portName)
			w.depth++
			w.line("Port: %s,", port)
			w.line("Protocol: %q,", "TCP")
			w.line("Endpoint: %q,", "/")
			w.line("ReadinessEndpoint: %q,", "/")
			w.line("LivenessEndpoint: %q,", "/")
			w.depth--
			w.line("},")
		}

		w.depth--
		w.line("},")
	}

	if len(flagEnvs) > 0 || len(flagExposes) > 0 {
		w.line("Env: map[string]deploy.EnvVar{")
		w.depth++

		for _, f := range flagEnvs {
			if f.Desc != "" {
				for line := range strings.Lines(toComment(f.Desc)) {
					w.line("%s", line)
				}
			}

			w.line("%q: {", f.EnvVar)
			w.depth++
			w.line("Value: %q,", f.DefaultValue())
			w.depth--
			w.line("},")
		}

		for _, f := range flagExposes {
			portName := f.Expose
			parts := strings.Split(f.String(), ":")
			port := parts[len(parts)-1]

			valueRef := f.String()
			if port != "" {
				placeholder := "{{ .Ports[" + strconv.Quote(portName) + "].Port }}"
				valueRef = strings.Replace(valueRef, port, placeholder, 1)
			}

			if f.Desc != "" {
				for line := range strings.Lines(toComment(f.Desc)) {
					w.line("%s", line)
				}
			}

			w.line("%q: {", f.EnvVar)
			w.depth++
			w.line("ValueRef: `%s`,", valueRef)
			w.depth--
			w.line("},")
		}

		w.depth--
		w.line("},")
	}

	w.depth--
	w.line("}")

	w.b.WriteByte('\n')
	w.line("return c")

	w.depth--
	w.line("}")

	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}

	out := w.b.String()

	// 用 gofmt 格式化生成的代码
	fmted, err := format.Source([]byte(out))
	if err != nil {
		// 格式化失败时仍写入原始内容，方便排查
		_ = os.WriteFile(path.Join(dest, "container.go"), []byte("// gofmt 格式化失败: "+err.Error()+"\n"+out), 0o600)
		return err
	}

	return os.WriteFile(path.Join(dest, "container.go"), fmted, 0o600)
}

// gofmtBuilder 辅助生成带一致缩进的 Go 代码。
type gofmtBuilder struct {
	b     strings.Builder
	depth int
}

func (w *gofmtBuilder) indent() {
	for i := 0; i < w.depth; i++ {
		w.b.WriteByte('\t')
	}
}

func (w *gofmtBuilder) line(format string, args ...any) {
	w.indent()
	fmt.Fprintf(&w.b, format, args...)
	w.b.WriteByte('\n')
}

func (w *gofmtBuilder) block(format string, args ...any) {
	w.indent()
	fmt.Fprintf(&w.b, format, args...)
	w.b.WriteString(" {\n")
	w.depth++
}

func (w *gofmtBuilder) endInline(format string, args ...any) {
	w.depth--
	w.indent()
	fmt.Fprintf(&w.b, format, args...)
	w.b.WriteString("\n")
}

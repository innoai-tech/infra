package workrun

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/mod/modfile"
)

type Runner struct {
	Args []string `arg:",interspersed"`

	Context string `flag:",omitzero" alias:"C"`
}

func (r *Runner) SetDefaults() {
	if r.Context == "" {
		r.Context = "./..."
	}
}

func (r *Runner) Run(ctx context.Context) error {
	if len(r.Args) == 0 {
		return fmt.Errorf("no command specified")
	}

	type target struct {
		Dir     string
		FullDir string
		Pattern string
	}

	var targets []target

	resolveArgs := func(args []string, pattern string) []string {
		resolved := make([]string, len(args))
		for i, arg := range args {
			resolved[i] = strings.ReplaceAll(arg, "{}", pattern)
		}
		return resolved
	}

	if r.Context == "./..." {
		workFile, err := findGoWork()
		if err != nil {
			return fmt.Errorf("find go.work: %w", err)
		}

		data, err := os.ReadFile(workFile)
		if err != nil {
			return fmt.Errorf("read go.work: %w", err)
		}

		wf, err := modfile.ParseWork(workFile, data, nil)
		if err != nil {
			return fmt.Errorf("parse go.work: %w", err)
		}

		workDir := filepath.Dir(workFile)

		for _, use := range wf.Use {
			dir := use.Path

			absDir := dir
			if !filepath.IsAbs(dir) {
				absDir = filepath.Join(workDir, dir)
			}

			targets = append(targets, target{
				Dir:     dir,
				FullDir: absDir,
				Pattern: "./...",
			})
		}
	} else {
		dir := r.Context
		pattern := "./"

		if cleaned, ok := strings.CutSuffix(dir, "/..."); ok {
			dir = cleaned
			pattern = "./..."
		}

		absDir := dir
		if !filepath.IsAbs(dir) {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			absDir = filepath.Join(cwd, dir)
		}

		targets = append(targets, target{
			Dir:     dir,
			FullDir: absDir,
			Pattern: pattern,
		})
	}

	for i, t := range targets {
		rr := &runner{
			Color:   dirColors[i%len(dirColors)],
			Dir:     t.Dir,
			FullDir: t.FullDir,
		}

		if err := rr.Exec(ctx, resolveArgs(r.Args, t.Pattern)); err != nil {
			return err
		}
	}

	return nil
}

var dirColors = []*color.Color{
	color.New(color.FgCyan),
	color.New(color.FgYellow),
	color.New(color.FgGreen),
	color.New(color.FgMagenta),
	color.New(color.FgBlue),
	color.New(color.FgHiCyan),
	color.New(color.FgHiYellow),
	color.New(color.FgHiGreen),
	color.New(color.FgHiMagenta),
	color.New(color.FgHiBlue),
}

type runner struct {
	Color   *color.Color
	Dir     string
	FullDir string
}

func (r *runner) Exec(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = r.FullDir
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	r.Infof("%s\n", strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		return err
	}

	wg := &sync.WaitGroup{}

	wg.Go(func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			r.Infof("%s\n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			r.Errorf("%s\n", err)
		}
	})

	wg.Go(func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			r.Errorf("%s\n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			r.Errorf("%s\n", err)
		}
	})

	wg.Wait()

	return cmd.Wait()
}

func (r *runner) Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s", r.Color.Sprintf("%s:", r.Dir))
	fmt.Fprintf(os.Stderr, format, args...)
}

func (r *runner) Infof(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s", r.Color.Sprintf("%s: ", r.Dir))
	fmt.Fprintf(os.Stdout, format, args...)
}

func findGoWork() (string, error) {
	if gw := os.Getenv("GOWORK"); gw != "" {
		if gw == "off" {
			return "", fmt.Errorf("GOWORK=off, no workspace")
		}
		return gw, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		p := filepath.Join(dir, "go.work")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.work not found")
}

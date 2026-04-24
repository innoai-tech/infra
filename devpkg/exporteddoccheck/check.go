package exporteddoccheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Finding 描述一个缺失导出注释的声明。
type Finding struct {
	File string
	Line int
	Kind string
	Name string
}

// CheckPackages 在给定根目录下扫描 Go package 并返回缺失导出注释的声明。
func CheckPackages(root string) ([]Finding, error) {
	dirs, err := packageDirs(root)
	if err != nil {
		return nil, err
	}

	findings := make([]Finding, 0)
	for _, dir := range dirs {
		list, err := CheckDir(dir)
		if err != nil {
			return nil, err
		}
		findings = append(findings, list...)
	}

	slices.SortFunc(findings, func(a, b Finding) int {
		if c := strings.Compare(a.File, b.File); c != 0 {
			return c
		}
		if a.Line != b.Line {
			return a.Line - b.Line
		}
		return strings.Compare(a.Name, b.Name)
	})

	return findings, nil
}

// CheckDir 检查单个目录中的 Go package。
func CheckDir(dir string) ([]Finding, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(info fs.FileInfo) bool {
		if info.IsDir() {
			return false
		}
		name := info.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return false
		}
		if strings.Contains(name, ".gen.") || strings.HasPrefix(name, "zz_generated.") {
			return false
		}
		return true
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse dir %s: %w", dir, err)
	}

	findings := make([]Finding, 0)
	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			if isGeneratedFile(file) {
				continue
			}
			findings = append(findings, checkFile(fset, filename, file)...)
		}
	}

	return findings, nil
}

func packageDirs(root string) ([]string, error) {
	dirs := make([]string, 0)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(root, entry.Name())
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			name := file.Name()
			if file.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			if strings.Contains(name, ".gen.") || strings.HasPrefix(name, "zz_generated.") {
				continue
			}
			dirs = append(dirs, path)
			break
		}
	}

	return dirs, nil
}

func isGeneratedFile(file *ast.File) bool {
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "Code generated") {
				return true
			}
		}
	}
	return false
}

func checkFile(fset *token.FileSet, filename string, file *ast.File) []Finding {
	findings := make([]Finding, 0)

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if !d.Name.IsExported() || hasDoc(d.Doc) || (d.Recv != nil && !exportedRecv(d.Recv)) {
				continue
			}
			findings = append(findings, newFinding(fset, filename, d.Pos(), "func", d.Name.Name))
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				for _, spec := range d.Specs {
					ts := spec.(*ast.TypeSpec)
					if !ts.Name.IsExported() {
						continue
					}
					if hasDoc(ts.Doc) || hasDoc(d.Doc) {
						continue
					}
					findings = append(findings, newFinding(fset, filename, ts.Pos(), "type", ts.Name.Name))
				}
			case token.CONST, token.VAR:
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					if hasDoc(vs.Doc) || hasDoc(d.Doc) {
						continue
					}
					for _, name := range vs.Names {
						if !name.IsExported() {
							continue
						}
						findings = append(findings, newFinding(fset, filename, name.Pos(), strings.ToLower(d.Tok.String()), name.Name))
					}
				}
			}
		}
	}

	return findings
}

func hasDoc(doc *ast.CommentGroup) bool {
	return doc != nil && strings.TrimSpace(doc.Text()) != ""
}

func exportedRecv(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) == 0 {
		return true
	}

	switch t := recv.List[0].Type.(type) {
	case *ast.Ident:
		return ast.IsExported(t.Name)
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ast.IsExported(ident.Name)
		}
	}

	return false
}

func newFinding(fset *token.FileSet, filename string, pos token.Pos, kind string, name string) Finding {
	p := fset.Position(pos)
	return Finding{
		File: filename,
		Line: p.Line,
		Kind: kind,
		Name: name,
	}
}

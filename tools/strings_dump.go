package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
)

type Row struct {
	File     string
	Line     int
	Function string
	Context  string
	Text     string
}

func main() {
	rows := []Row{}
	for _, name := range os.Args[1:] {
		fs := token.NewFileSet()
		f, e := parser.ParseFile(fs, name, nil, 0)
		if e != nil {
			panic(e)
		}
		for _, d := range f.Decls {
			fn := ""
			if x, ok := d.(*ast.FuncDecl); ok {
				fn = x.Name.Name
			}
			stack := []ast.Node{}
			ast.Inspect(d, func(n ast.Node) bool {
				if n == nil {
					stack = stack[:len(stack)-1]
					return true
				}
				if x, ok := n.(*ast.BasicLit); ok && x.Kind == token.STRING {
					s, _ := strconv.Unquote(x.Value)
					context := "literal"
					for _, parent := range stack {
						if _, ok := parent.(*ast.ImportSpec); ok {
							context = "import"
						}
					}
					for i := len(stack) - 1; i >= 0; i-- {
						if call, ok := stack[i].(*ast.CallExpr); ok {
							var b bytes.Buffer
							printer.Fprint(&b, fs, call.Fun)
							arg := -1
							for j, a := range call.Args {
								if a.Pos() <= x.Pos() && a.End() >= x.End() {
									arg = j
									break
								}
							}
							context = b.String() + ":" + strconv.Itoa(arg)
							break
						}
					}
					rows = append(rows, Row{filepath.Base(name), fs.Position(x.Pos()).Line, fn, context, s})
				}
				stack = append(stack, n)
				return true
			})
		}
	}
	json.NewEncoder(os.Stdout).Encode(rows)
}

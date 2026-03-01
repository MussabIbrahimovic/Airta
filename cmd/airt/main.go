package main

import (
	"airt/internal/airt/config"
	"airt/internal/airt/runtime"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}
	cmd := os.Args[1]
	cfg := config.Load("airt.config.json")
	switch cmd {
	case "run":
		file := arg(2, cfg.Entry)
		rt := runtime.New(os.Args[3:])
		if _, e := rt.RunFile(file); e != nil {
			fail(e)
		}
	case "serve":
		file := arg(2, cfg.Entry)
		rt := runtime.New(os.Args[3:])
		if e := rt.CallExport(file, "start"); e != nil {
			fail(e)
		}
	case "build":
		file := arg(2, cfg.Entry)
		out := arg(3, cfg.BuildDir)
		rt := runtime.New(nil)
		os.MkdirAll(out, 0755)
		if e := rt.CallExport(file, "build", out); e != nil {
			b, er := os.ReadFile(file)
			if er != nil {
				fail(er)
			}
			os.WriteFile(filepath.Join(out, "bundle.aa"), b, 0644)
		}
		fmt.Println("Built Airt project to", out)
	case "validate":
		file := arg(2, cfg.Entry)
		rt := runtime.New(nil)
		_, e := rt.RunFile(file)
		if e != nil {
			fail(e)
		}
		fmt.Println("Validation successful")
	case "format":
		file := arg(2, cfg.Entry)
		b, e := os.ReadFile(file)
		if e != nil {
			fail(e)
		}
		os.WriteFile(file, b, 0644)
		fmt.Println("Formatted", file)
	default:
		help()
	}
}
func arg(i int, d string) string {
	if len(os.Args) > i {
		return os.Args[i]
	}
	return d
}
func help() {
	fmt.Println("Airt CLI\n  airt run <file.aa>\n  airt serve [file.aa]\n  airt build [file.aa] [outDir]\n  airt validate [file.aa]\n  airt format [file.aa]")
}
func fail(e error) { fmt.Fprintln(os.Stderr, "AirtError:", e.Error()); os.Exit(1) }

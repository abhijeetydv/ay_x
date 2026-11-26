package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"portfolio/internal/loader"
	"portfolio/internal/render"
	"portfolio/internal/validate"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	contentDir := fs.String("content", "content", "path to YAML content directory")
	templatePath := fs.String("template", filepath.Join("templates", "home.html"), "path to homepage template")
	outDir := fs.String("out", "dist", "path to output directory")
	_ = fs.Parse(os.Args[2:])

	switch cmd {
	case "build":
		if err := runBuild(*contentDir, *templatePath, *outDir); err != nil {
			fmt.Fprintf(os.Stderr, "build error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("built homepage in %s/\n", *outDir)
	case "validate":
		if err := runValidate(*contentDir); err != nil {
			fmt.Fprintf(os.Stderr, "validation error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("all YAML content is valid")
	default:
		usage()
		os.Exit(1)
	}
}

func runValidate(contentDir string) error {
	data, err := loader.LoadAll(contentDir)
	if err != nil {
		return err
	}
	return validate.Homepage(data)
}

func runBuild(contentDir, templatePath, outDir string) error {
	data, err := loader.LoadAll(contentDir)
	if err != nil {
		return err
	}
	if err := validate.Homepage(data); err != nil {
		return err
	}
	return render.BuildHomepage(templatePath, ".", outDir, data)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./cmd/portfolio <build|validate> [-content dir] [-template file] [-out dir]")
}

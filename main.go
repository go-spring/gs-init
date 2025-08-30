package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const Version = "v0.0.1"

func main() {
	var (
		version bool
		module  string
		branch  string
	)

	root := &cobra.Command{
		Use:          "gs-init",
		Short:        "init go server project",
		SilenceUsage: true,
	}

	root.Flags().BoolVar(&version, "version", false, "show version")
	root.Flags().StringVar(&module, "module", "", "module name, required")
	root.Flags().StringVar(&branch, "branch", "main", "git branch")

	root.RunE = func(cmd *cobra.Command, args []string) error {
		if version {
			fmt.Println(root.Short)
			fmt.Println(Version)
			return nil
		}

		if module == "" {
			log.Fatalln("module name is required")
		}

		ss := strings.Split(module, "/")
		projectName := ss[len(ss)-1]

		if _, err := os.Stat(projectName); err != nil {
			if !os.IsNotExist(err) {
				log.Fatalln(err)
			}
		} else {
			log.Fatalln("project already exists")
		}

		srcDir := gitClone(branch)
		fmt.Println(srcDir)

		pkgName := toPascal(projectName)
		replaceFiles(srcDir, module, pkgName)

		if err := os.Rename(srcDir, projectName); err != nil {
			log.Fatalln(err)
		}

		return nil
	}

	if err := root.Execute(); err != nil {
		os.Exit(-1)
	}
}

func gitClone(branch string) string {
	tempDir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		log.Fatalln(err)
	}
	cmd := exec.Command(
		"git",
		"clone",
		"--depth",
		"1",
		"--branch",
		branch,
		"--single-branch",
		"https://github.com/go-spring/skeleton.git",
	)
	cmd.Dir = tempDir
	cmd.Env = os.Environ()
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalln(err, string(b))
	}
	log.Println(string(b))
	projectDir := filepath.Join(tempDir, "skeleton")
	gitDir := filepath.Join(projectDir, ".git")
	if err = os.RemoveAll(gitDir); err != nil {
		log.Fatalln(err)
	}
	return projectDir
}

func toPascal(s string) string {
	var sb strings.Builder
	parts := strings.Split(s, "_")
	for _, part := range parts {
		if part == "" {
			continue
		}
		c := part[0]
		if 'a' <= c && c <= 'z' {
			c = c - 'a' + 'A'
		}
		sb.WriteByte(c)
		if len(part) > 1 {
			sb.WriteString(part[1:])
		}
	}
	return sb.String()
}

func replaceFiles(dir string, module, pkgName string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalln(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			replaceFiles(filepath.Join(dir, e.Name()), module, pkgName)
			continue
		}
		file := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(file)
		if err != nil {
			log.Fatalln(err)
		}

		b = bytes.ReplaceAll(b, []byte("GS_PROJECT_MODULE"), []byte(module))
		b = bytes.ReplaceAll(b, []byte("GS_PROJECT_NAME"), []byte(pkgName))

		if err = os.Remove(file); err != nil {
			log.Fatalln(err)
		}

		file = strings.ReplaceAll(file, "GS_PROJECT_NAME", pkgName)
		if err = os.WriteFile(file, b, os.ModePerm); err != nil {
			log.Fatalln(err)
		}
	}
}

/*
 * Copyright 2025 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

		// Extract project name from module path
		ss := strings.Split(module, "/")
		projectName := ss[len(ss)-1]

		// Check if project directory already exists
		if _, err := os.Stat(projectName); err != nil {
			if !os.IsNotExist(err) {
				log.Fatalln(err)
			}
		} else {
			log.Fatalln("project already exists")
		}

		// Clone the skeleton repository
		srcDir := gitClone(branch)
		fmt.Println(srcDir)

		// Convert project name to PascalCase for Go package naming
		pkgName := toPascal(projectName)
		replaceFiles(srcDir, module, pkgName)

		// Rename project directory
		if err := os.Rename(srcDir, projectName); err != nil {
			log.Fatalln(err)
		}

		return nil
	}

	if err := root.Execute(); err != nil {
		os.Exit(-1)
	}
}

// gitClone clones the skeleton project to a temporary directory
func gitClone(branch string) string {
	tempDir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(tempDir) // log temp directory path

	// Execute git clone
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
	log.Println(string(b)) // output git clone result

	// Remove .git folder to detach from skeleton repo
	projectDir := filepath.Join(tempDir, "skeleton")
	gitDir := filepath.Join(projectDir, ".git")
	if err = os.RemoveAll(gitDir); err != nil {
		log.Fatalln(err)
	}
	return projectDir
}

// toPascal converts snake_case string to PascalCase
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

// replaceFiles recursively replaces placeholders in files and filenames
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

		fileName := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(fileName)
		if err != nil {
			log.Fatalln(err)
		}

		// Replace placeholders in file content
		b = bytes.ReplaceAll(b, []byte("GS_PROJECT_MODULE"), []byte(module))
		b = bytes.ReplaceAll(b, []byte("GS_PROJECT_NAME"), []byte(pkgName))

		// Remove original file (preparing to rename if necessary)
		if err = os.Remove(fileName); err != nil {
			log.Fatalln(err)
		}

		// Write updated content to file
		fileName = strings.ReplaceAll(fileName, "GS_PROJECT_NAME", pkgName)
		if err = os.WriteFile(fileName, b, os.ModePerm); err != nil {
			log.Fatalln(err)
		}
	}
}

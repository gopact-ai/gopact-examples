package repositorychecks

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamplesDocumentationFilesStayUnderDocExceptReadmes(t *testing.T) {
	err := filepath.WalkDir("../..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}

		rel, err := filepath.Rel("../..", path)
		if err != nil {
			return err
		}
		slashPath := filepath.ToSlash(rel)
		if entry.Name() == "README.md" || entry.Name() == "README_zh.md" || strings.HasPrefix(slashPath, "doc/") {
			return nil
		}
		t.Fatalf("%s is a Markdown document outside doc/ and is not a README", slashPath)
		return nil
	})
	if err != nil {
		t.Fatalf("walk markdown docs: %v", err)
	}
}

func TestExamplesMarkdownDocsUseSeparatedLanguageFiles(t *testing.T) {
	err := filepath.WalkDir("../..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}

		body := readText(t, path)
		slashPath := filepath.ToSlash(path)
		if strings.HasSuffix(entry.Name(), "_zh.md") {
			if !strings.Contains(body, "<!-- gopact:doc-language: zh -->") {
				t.Fatalf("%s missing Chinese documentation marker", slashPath)
			}
			if strings.Contains(body, "## English") {
				t.Fatalf("%s must not embed English content", slashPath)
			}
			return nil
		}
		if !strings.Contains(body, "<!-- gopact:doc-language: en -->") {
			t.Fatalf("%s missing English documentation marker", slashPath)
		}
		if strings.Contains(body, "## 中文") {
			t.Fatalf("%s must not embed Chinese content", slashPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk markdown docs: %v", err)
	}
}

func TestExamplesReadmeBadgesAndDocIndexAreConfigured(t *testing.T) {
	readme := readText(t, "../../README.md")

	for _, want := range []string{
		"https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main",
		"https://img.shields.io/github/license/gopact-ai/gopact-examples",
		"https://pkg.go.dev/badge/github.com/gopact-ai/gopact-examples.svg",
		"doc/README.md",
		"doc/FEATURES.md",
		"doc/CONTRIBUTING.md",
		"doc/SECURITY.md",
		"doc/CHANGELOG.md",
		"doc/maintainers/repository-governance.md",
	} {
		if !strings.Contains(readme, want) {
			t.Fatalf("README.md missing badge or doc index entry %q", want)
		}
	}
}

// Package transpiler validates Scala applications and generates Scala CLI scripts.
package transpiler

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Transpiler holds the supported Scala versions and fixed dependencies.
type Transpiler struct {
	defaultVersion string
	versions       []string
	dependencies   []string
}

// New constructs a transpiler from the application's fixed configuration.
func New(defaultVersion string, versions, dependencies []string) *Transpiler {
	return &Transpiler{defaultVersion: defaultVersion, versions: slices.Clone(versions), dependencies: slices.Clone(dependencies)}
}

// Input contains source code and the requested Scala version.
type Input struct{ Code, ScalaVersion string }

// ValidatedInput can only be populated by Validate.
type ValidatedInput struct{ code, version, entry string }

// Script is the generated Scala CLI file.
type Script struct {
	Filename string `json:"filename"`
	Source   string `json:"source"`
}

// Validate resolves defaults and checks input before script generation.
func (t *Transpiler) Validate(input Input) (ValidatedInput, error) {
	if strings.TrimSpace(input.Code) == "" {
		return ValidatedInput{}, errors.New("コードを入力してください。")
	}
	entry, err := applicationEntry(input.Code)
	if err != nil {
		return ValidatedInput{}, err
	}
	version := input.ScalaVersion
	if version == "" {
		version = t.defaultVersion
	}
	if !slices.Contains(t.versions, version) {
		return ValidatedInput{}, errors.New("選択できない Scala バージョンです。")
	}
	return ValidatedInput{code: input.Code, version: version, entry: entry}, nil
}

// Transpile renders a previously validated application without executing it.
func (t *Transpiler) Transpile(input ValidatedInput) Script {
	var header strings.Builder
	fmt.Fprintf(&header, "//> using scala %s\n", input.version)
	for _, coordinate := range t.dependencies {
		fmt.Fprintf(&header, "//> using dep %s\n", coordinate)
	}
	return Script{Filename: "playground.sc", Source: header.String() + "\n" + input.code + "\n\n" + input.entry + ".main(Array.empty[String])\n"}
}

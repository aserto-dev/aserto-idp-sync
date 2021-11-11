//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/aserto-dev/mage-loot/common"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

func init() {
	// Set go version for docker builds
	os.Setenv("GO_VERSION", "1.16")
	// Set private repositories
	os.Setenv("GOPRIVATE", "github.com/aserto-dev")
	// Enable docker buildkit capabilities
	os.Setenv("DOCKER_BUILDKIT", "1")
}

// Generate generates all code.
func Generate() error {
	// These extra commands are required because of
	// https://github.com/golang/go/issues/44129

	if err := sh.RunV("buf", "generate"); err != nil {
		return err
	}
	if err := sh.RunV("go", "get", "-tags", "wireinject", "./..."); err != nil {
		return err
	}
	if err := sh.RunV("go", "install", "github.com/google/wire/cmd/wire"); err != nil {
		return err
	}
	if err := sh.RunV("go", "mod", "download"); err != nil {
		return err
	}
	if err := common.Generate(); err != nil {
		return err
	}
	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return err
	}

	return nil
}

// Build builds all binaries in ./cmd.
func Build() error {
	return common.BuildReleaser()
}

// BuildAll builds all binaries in ./cmd for
// all configured operating systems and architectures.
func BuildAll() error {
	return common.BuildAllReleaser()
}

// Lint runs linting for the entire project.
func Lint() error {
	return common.Lint()
}

// Test runs all tests and generates a code coverage report.
func Test() error {
	return common.Test()
}

// DockerBuild builds the docker image for the project.
func DockerBuild() error {
	version, err := common.Version()
	if err != nil {
		return errors.Wrap(err, "failed to calculate version")
	}

	return common.DockerImage(fmt.Sprintf("aserto-idp-sync:%s", version), "--platform=linux/amd64")
}

// Deps installs all tool dependencies.
func Deps() {
	deps.GetAllDeps()
}

// All runs all targets in the appropriate order.
// The targets are run in the following order:
// deps, generate, lint, test, build, dockerImage
func All() error {
	mg.SerialDeps(Deps, Generate, Lint, Test, Build, DockerBuild)
	return nil
}

func ldflags() ([]string, error) {
	commit, err := common.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate git commit")
	}
	version, err := common.Version()
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate version")
	}

	date := time.Now().UTC().Format(time.RFC3339)

	ldbase := "github.com/aserto-dev/aserto-idp-sync/pkg/version"
	ldflags := fmt.Sprintf(`"-X %s.ver=%s -X %s.commit=%s -X %s.date=%s"`,
		ldbase, commit, ldbase, version, ldbase, date)

	return []string{"-ldflags", ldflags}, nil
}

// Release releases the project.
func Release() error {
	return common.Release()
}

func Run() error {
	return sh.RunV(
		"./dist/aserto-idp-sync_"+runtime.GOOS+"_"+runtime.GOARCH+"/aserto-idp-sync",
		"run",
		"--config", "./pkg/testharness/testdata/config.yaml",
	)
}

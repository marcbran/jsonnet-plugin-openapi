//go:build e2e

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
)

type CLIFacade struct {
	binaryPath string
}

func NewCLIFacade() (openapipkg.Facade, error) {
	binaryPath := os.Getenv("JSONNET_OPENAPI_BINARY")
	if binaryPath == "" {
		return nil, fmt.Errorf("JSONNET_OPENAPI_BINARY is required")
	}
	_, err := os.Stat(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("unable to access JSONNET_OPENAPI_BINARY %q: %w", binaryPath, err)
	}
	return &CLIFacade{binaryPath: binaryPath}, nil
}

func (f *CLIFacade) Generate(ctx context.Context, in openapipkg.Input) (openapipkg.Output, error) {
	args := []string{
		"gen",
		in.Spec,
		"--out",
		in.OutDir,
		"--format",
		"json",
		"--service",
		in.Service,
	}
	if in.PkgRepo != "" {
		args = append(args, "--pkg-repo", in.PkgRepo)
	}
	args = append(args, "-q")
	cmd := exec.CommandContext(ctx, f.binaryPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if stderr.String() != "" {
			return openapipkg.Output{}, errors.New(stderr.String())
		}
		return openapipkg.Output{}, err
	}
	var out openapipkg.Output
	err = json.Unmarshal(stdout.Bytes(), &out)
	if err != nil {
		return openapipkg.Output{}, err
	}
	return out, nil
}

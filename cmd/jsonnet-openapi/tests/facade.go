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

func (f *CLIFacade) Batch(ctx context.Context, jobs []openapipkg.Input) ([]openapipkg.Output, error) {
	fh, err := os.CreateTemp("", "openapi-batch-*.json")
	if err != nil {
		return nil, err
	}
	path := fh.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}
	err = json.NewEncoder(fh).Encode(jobs)
	if err != nil {
		cleanup()
		_ = fh.Close()
		return nil, err
	}
	err = fh.Close()
	if err != nil {
		cleanup()
		return nil, err
	}
	defer cleanup()
	args := []string{
		"batch",
		path,
		"--format",
		"json",
		"-q",
	}
	cmd := exec.CommandContext(ctx, f.binaryPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		if stderr.String() != "" {
			return nil, errors.New(stderr.String())
		}
		return nil, err
	}
	var outs []openapipkg.Output
	err = json.Unmarshal(stdout.Bytes(), &outs)
	if err != nil {
		return nil, err
	}
	return outs, nil
}

func (f *CLIFacade) Generate(ctx context.Context, in openapipkg.Input) (openapipkg.Output, error) {
	args := []string{
		"gen",
		in.Ref,
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

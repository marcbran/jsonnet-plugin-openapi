package listcolumns

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/inference"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/infra/codex"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/infra/files"
	infrajsonnet "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/infra/jsonnet"
	infrakinopenapi "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/infra/kinopenapi"
)

const defaultModel = "gpt-5.5"

type Input struct {
	Spec    string
	Out     string
	WorkDir string
	Model   string
	Limit   int
	Force   bool
}

type Output struct {
	Out     string   `json:"out"`
	WorkDir string   `json:"workDir"`
	Files   []string `json:"files"`
}

func Exec(ctx context.Context, in Input) (Output, error) {
	specDir, specName, err := localOutputDefaults(in.Spec)
	if err != nil {
		return Output{}, err
	}
	workDir := in.WorkDir
	if workDir == "" {
		workDir = filepath.Join(specDir, specName)
	}
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return Output{}, err
	}
	out := in.Out
	if out == "" {
		out = filepath.Join(specDir, specName+".columns.json")
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return Output{}, err
	}
	model := in.Model
	if model == "" {
		model = defaultModel
	}

	loader := infrakinopenapi.NewLoader()
	lib, err := Lib()
	if err != nil {
		return Output{}, err
	}
	sharedLib, err := inference.Lib()
	if err != nil {
		return Output{}, err
	}
	renderer := infrajsonnet.NewRenderer(lib, sharedLib)
	store := files.NewStore(workDir)
	pipeline := inference.Pipeline{
		Jobs: []inference.Job{
			NewColumnsJob(renderer),
		},
		Runner: codex.NewRunner(model),
		Store:  store,
		Force:  in.Force,
		Limit:  in.Limit,
	}

	spec, err := loader.LoadSpec(ctx, in.Spec)
	if err != nil {
		return Output{}, err
	}
	results, err := pipeline.Exec(ctx, spec)
	if err != nil {
		return Output{}, err
	}
	columns, err := renderer.RenderOutput("list-columns.jsonnet",
		infrajsonnet.Binding{Name: "spec", Value: spec.JSON},
		infrajsonnet.Binding{Name: "inferred", Value: results[ColumnsJobName]},
	)
	if err != nil {
		return Output{}, err
	}

	err = os.MkdirAll(filepath.Dir(out), 0755)
	if err != nil {
		return Output{}, err
	}
	err = os.WriteFile(out, columns, 0644)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Out:     out,
		WorkDir: workDir,
		Files: []string{
			relativeOrAbs(workDir, filepath.Join(workDir, ColumnsJobName, "results", "all.jsonnet")),
			out,
		},
	}, nil
}

func localOutputDefaults(ref string) (dir string, name string, err error) {
	if isHTTPRef(ref) {
		u, err := url.Parse(ref)
		if err != nil {
			return "", "", err
		}
		name = strings.TrimSuffix(filepath.Base(u.Path), filepath.Ext(u.Path))
		if name == "" || name == "." || name == "/" {
			name = "openapi"
		}
		dir, err = os.Getwd()
		if err != nil {
			return "", "", err
		}
		return dir, name, nil
	}
	abs, err := filepath.Abs(ref)
	if err != nil {
		return "", "", err
	}
	name = strings.TrimSuffix(filepath.Base(abs), filepath.Ext(abs))
	if name == "" {
		name = "openapi"
	}
	return filepath.Dir(abs), name, nil
}

func isHTTPRef(ref string) bool {
	return strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://")
}

func relativeOrAbs(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

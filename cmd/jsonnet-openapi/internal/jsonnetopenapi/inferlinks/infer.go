package inferlinks

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/marcbran/jpoet/pkg/jpoet"
)

const defaultModel = "gpt-5.5"

const listDetailPrompt = `Read input.json.

Infer whether the OpenAPI list response at sourcePath has a canonical GET detail path among detailPaths.

Return only JSON matching the provided schema.

Rules:
- Choose "detail_elsewhere" only when there is a clear canonical detail GET endpoint for the list item type.
- Choose "no_detail_get" when the list item is an event, stats/summary object, search result, relationship record, activity feed item, or otherwise has no canonical detail GET path in detailPaths.
- Choose "uncertain" when there is not enough evidence.
- For "detail_elsewhere", targetPath must be one path from detailPaths and array must be the array path from input.json.
- For "no_detail_get" or "uncertain", targetPath must be null.
- Do not invent paths.
- Do not infer variable mappings.
`

const listDetailVarsPrompt = `Read input.json.

Infer JSON property paths on the array item that provide values for the target path params listed in missingParams.

Return only JSON matching the provided schema.

Rules:
- Only infer vars for params in missingParams.
- Each vars value must be a property path relative to the array item, for example ["account", "id"] or ["name"].
- Do not include params that are already present in inheritedParams.
- Do not invent properties that are not supported by itemSchema.
- Match the meaning of each target path param, not just its name.
- Prefer stable canonical identifiers over display names.
- Prefer exact or clearly equivalent property names when available, for example an "id" param from an "id" property, or a "name" param from a "name" property.
- For params ending in "_id" or named "id", prefer stable id-like fields over names, slugs, titles, or URLs.
- For params ending in "_name" or named "name", prefer stable name-like fields over display titles or descriptions.
- For slug/key/code params, prefer slug/key/code fields over human-readable labels.
- Avoid URLs, descriptions, titles, summaries, display names, timestamps, booleans, counts, and status fields unless the target param clearly asks for that value.
- Return vars as an array of objects with param and path fields.
- If a missing param cannot be resolved from itemSchema, omit it from vars and explain that in reason.
`

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

func Run(ctx context.Context, in Input) (Output, error) {
	spec, err := filepath.Abs(in.Spec)
	if err != nil {
		return Output{}, err
	}
	specJSON, err := os.ReadFile(spec)
	if err != nil {
		return Output{}, err
	}
	specDir := filepath.Dir(spec)
	specName := strings.TrimSuffix(filepath.Base(spec), filepath.Ext(spec))
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
		out = filepath.Join(specDir, specName+".links.json")
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return Output{}, err
	}
	model := in.Model
	if model == "" {
		model = defaultModel
	}

	sub, err := fs.Sub(lib, "lib")
	if err != nil {
		return Output{}, err
	}

	listDetailDir := filepath.Join(workDir, "list-detail-inference")
	err = generateBundles(sub, string(specJSON), "", "list-detail-inference-bundles.jsonnet", filepath.Join(listDetailDir, "bundles"))
	if err != nil {
		return Output{}, err
	}
	listDetailResults := filepath.Join(listDetailDir, "results")
	err = runCodexStep(ctx, codexStep{
		WorkDir:    listDetailDir,
		ResultsDir: listDetailResults,
		SchemaName: "list-detail-inference-output.schema.json",
		Prompt:     listDetailPrompt,
		Model:      model,
		Limit:      in.Limit,
		Force:      in.Force,
	})
	if err != nil {
		return Output{}, err
	}
	inferred, err := writeImportsAndReadResults(listDetailResults)
	if err != nil {
		return Output{}, err
	}

	varsDir := filepath.Join(workDir, "list-detail-vars-inference")
	err = generateBundles(sub, string(specJSON), inferred, "list-detail-vars-inference-bundles.jsonnet", filepath.Join(varsDir, "bundles"))
	if err != nil {
		return Output{}, err
	}
	varsResults := filepath.Join(varsDir, "results")
	err = runCodexStep(ctx, codexStep{
		WorkDir:    varsDir,
		ResultsDir: varsResults,
		SchemaName: "list-detail-vars-inference-output.schema.json",
		Prompt:     listDetailVarsPrompt,
		Model:      model,
		Limit:      in.Limit,
		Force:      in.Force,
	})
	if err != nil {
		return Output{}, err
	}
	varsInferred, err := writeImportsAndReadResults(varsResults)
	if err != nil {
		return Output{}, err
	}

	err = os.MkdirAll(filepath.Dir(out), 0755)
	if err != nil {
		return Output{}, err
	}
	fh, err := os.Create(out)
	if err != nil {
		return Output{}, err
	}
	defer fh.Close()
	err = jpoet.Eval(
		jpoet.FSImport(sub),
		jpoet.SnippetInput("list-detail-links", fmt.Sprintf(
			"local spec = %s; local inferred = %s; local varsInferred = %s; (import 'list-detail-links.jsonnet')(spec, inferred, varsInferred)",
			string(specJSON),
			inferred,
			varsInferred,
		)),
		jpoet.WriterOutput(fh),
	)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Out:     out,
		WorkDir: workDir,
		Files: []string{
			relativeOrAbs(workDir, filepath.Join(listDetailResults, "all.jsonnet")),
			relativeOrAbs(workDir, filepath.Join(varsResults, "all.jsonnet")),
			out,
		},
	}, nil
}

func generateBundles(imports fs.FS, specJSON, inferred, file, outDir string) error {
	err := os.RemoveAll(outDir)
	if err != nil {
		return err
	}
	opts := []jpoet.Option{
		jpoet.FSImport(imports),
		jpoet.SnippetInput(file, fmt.Sprintf("local spec = %s; (import %q)(spec)", specJSON, file)),
		jpoet.Serialize(false),
		jpoet.DirectoryOutput(outDir),
	}
	if inferred != "" {
		opts[1] = jpoet.SnippetInput(file, fmt.Sprintf(
			"local spec = %s; local inferred = %s; (import %q)(spec, inferred)",
			specJSON,
			inferred,
			file,
		))
	}
	return jpoet.Eval(opts...)
}

type codexStep struct {
	WorkDir    string
	ResultsDir string
	SchemaName string
	Prompt     string
	Model      string
	Limit      int
	Force      bool
}

func runCodexStep(ctx context.Context, step codexStep) error {
	err := os.MkdirAll(step.ResultsDir, 0755)
	if err != nil {
		return err
	}
	bundlesDir := filepath.Join(step.WorkDir, "bundles")
	schemaPath := filepath.Join(step.WorkDir, step.SchemaName)
	schema, err := fs.ReadFile(lib, filepath.Join("lib", step.SchemaName))
	if err != nil {
		return err
	}
	err = os.WriteFile(schemaPath, schema, 0644)
	if err != nil {
		return err
	}
	promptPath := filepath.Join(step.WorkDir, "prompt.md")
	err = os.WriteFile(promptPath, []byte(step.Prompt), 0644)
	if err != nil {
		return err
	}
	inputs, err := filepath.Glob(filepath.Join(bundlesDir, "*", "input.json"))
	if err != nil {
		return err
	}
	sort.Strings(inputs)
	count := 0
	for _, input := range inputs {
		bundleDir := filepath.Dir(input)
		bundle := filepath.Base(bundleDir)
		output := filepath.Join(step.ResultsDir, bundle+".json")
		if !step.Force && fileExists(output) {
			continue
		}
		err = os.WriteFile(filepath.Join(bundleDir, "prompt.md"), []byte(step.Prompt), 0644)
		if err != nil {
			return err
		}
		err = runCodex(ctx, bundleDir, schemaPath, output, step.Model)
		if err != nil {
			return err
		}
		count++
		if step.Limit > 0 && count >= step.Limit {
			break
		}
	}
	return nil
}

func runCodex(ctx context.Context, dir, schema, output, model string) error {
	args := []string{
		"exec",
		"--cd", dir,
		"--skip-git-repo-check",
		"--ephemeral",
		"--sandbox", "read-only",
		"--model", model,
		"--output-schema", schema,
		"--output-last-message", output,
		"Read prompt.md and input.json. Return only JSON.",
	}
	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Stdin = strings.NewReader("")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("codex exec failed for %s: %w: %s", dir, err, msg)
		}
		return fmt.Errorf("codex exec failed for %s: %w", dir, err)
	}
	return nil
}

func writeImportsAndReadResults(resultsDir string) (string, error) {
	files, err := filepath.Glob(filepath.Join(resultsDir, "*.json"))
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	var imports strings.Builder
	imports.WriteString("[\n")
	var results strings.Builder
	results.WriteString("[\n")
	for _, file := range files {
		base := filepath.Base(file)
		imports.WriteString(fmt.Sprintf("  import %q,\n", base))
		raw, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		results.WriteString(string(bytes.TrimSpace(raw)))
		results.WriteString(",\n")
	}
	imports.WriteString("]\n")
	results.WriteString("]\n")
	err = os.WriteFile(filepath.Join(resultsDir, "all.jsonnet"), []byte(imports.String()), 0644)
	if err != nil {
		return "", err
	}
	return results.String(), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func relativeOrAbs(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

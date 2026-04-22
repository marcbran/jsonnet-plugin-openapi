package jsonnetopenapi

import (
	"encoding/json"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-jsonnet/jsonnet"
	"github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/internal/jsonnetopenapi/lib/imports"
)

func writeGeneratedLibsonnet(outDir string, payload *GenPayload) error {
	apiJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	err = jpoet.Eval(
		jpoet.FileImport([]string{}),
		jpoet.FSImport(lib),
		jpoet.FSImport(imports.Fs),
		jpoet.WithPlugin(jsonnet.Plugin()),
		jpoet.TLACode("api", string(apiJSON)),
		jpoet.FileInput("./lib/gen.libsonnet"),
		jpoet.Serialize(false),
		jpoet.DirectoryOutput(outDir),
	)
	if err != nil {
		return err
	}
	return nil
}

type GenPayload struct {
	Info    GenInfo   `json:"info"`
	Service string    `json:"service"`
	PkgRepo string    `json:"pkgRepo"`
	Trie    *TrieNode `json:"trie"`
}

type GenInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type TrieNode struct {
	Leaf     *GenOperation        `json:"leaf,omitempty"`
	Children map[string]*TrieNode `json:"children,omitempty"`
}

type GenOperation struct {
	ID           string      `json:"id"`
	PathTemplate string      `json:"pathTemplate"`
	PathFormat   string      `json:"pathFormat"`
	PathArgNames []string    `json:"pathArgNames"`
	QueryParams  []ParamSpec `json:"queryParams"`
	HeaderParams []ParamSpec `json:"headerParams"`
}

type ParamSpec struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

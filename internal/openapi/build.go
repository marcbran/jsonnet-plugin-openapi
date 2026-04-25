package openapi

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

var pathTplVar = regexp.MustCompile(`\{([^}]+)}`)

func BuildNestedSpec(api APISpec) (*NestedSpec, error) {
	root := &PathNode{}
	ops := collectGETOperations(api.Paths)
	count := 0
	for _, op := range ops {
		spec, err := operationFromGET(op)
		if err != nil {
			return nil, err
		}
		segs := pathSegments(op.Path)
		err = root.insert(segs, spec)
		if err != nil {
			return nil, err
		}
		count++
	}
	if count == 0 {
		return nil, fmt.Errorf("no GET operations in spec")
	}
	return &NestedSpec{
		Title:   api.Title,
		Version: api.Version,
		Paths:   root,
	}, nil
}

func collectGETOperations(paths []PathItem) []GETOperation {
	out := make([]GETOperation, 0, len(paths))
	for _, p := range paths {
		if p.Get == nil {
			continue
		}
		out = append(out, GETOperation{
			Path:        p.Path,
			OperationID: p.Get.OperationID,
			Parameters:  canonicalizeParameters(mergeParameters(p.Parameters, p.Get.Parameters)),
		})
	}
	return out
}

type GETOperation struct {
	Path        string
	OperationID string
	Parameters  []Parameter
}

func mergeParameters(pathParams, opParams []Parameter) []Parameter {
	byKey := make(map[string]Parameter, len(pathParams)+len(opParams))
	for _, p := range pathParams {
		byKey[p.In+":"+p.Name] = p
	}
	for _, p := range opParams {
		byKey[p.In+":"+p.Name] = p
	}
	out := make([]Parameter, 0, len(byKey))
	for _, p := range byKey {
		out = append(out, p)
	}
	return out
}

func canonicalizeParameters(params []Parameter) []Parameter {
	if len(params) <= 1 {
		return params
	}
	out := append([]Parameter(nil), params...)
	sort.Slice(out, func(i, j int) bool {
		ki := out[i].In + ":" + out[i].Name
		kj := out[j].In + ":" + out[j].Name
		return ki < kj
	})
	return out
}

func operationFromGET(op GETOperation) (*OperationSpec, error) {
	id := op.OperationID
	if id == "" {
		id = "get_" + slugify(strings.Trim(op.Path, "/"))
	}
	pf, pargs := pathFormatFromTemplate(op.Path)
	qp, hp := splitParams(op.Parameters)
	return &OperationSpec{
		ID:           id,
		PathTemplate: op.Path,
		PathFormat:   pf,
		PathArgNames: pargs,
		QueryParams:  qp,
		HeaderParams: hp,
	}, nil
}

func pathFormatFromTemplate(path string) (format string, names []string) {
	names = []string{}
	rest := path
	for {
		loc := pathTplVar.FindStringIndex(rest)
		if loc == nil {
			format += rest
			return format, names
		}
		format += rest[:loc[0]]
		format += "%s"
		sub := pathTplVar.FindStringSubmatch(rest[loc[0]:])[1]
		sub = strings.TrimPrefix(sub, ".")
		sub = strings.TrimSuffix(sub, "*")
		names = append(names, sub)
		rest = rest[loc[1]:]
	}
}

func splitParams(params []Parameter) (query []ParameterSpec, header []ParameterSpec) {
	query = []ParameterSpec{}
	header = []ParameterSpec{}
	for _, p := range params {
		ps := ParameterSpec{Name: p.Name, Required: p.Required}
		switch p.In {
		case "query":
			query = append(query, ps)
		case "header":
			header = append(header, ps)
		}
	}
	return query, header
}

func pathSegments(path string) []string {
	p := strings.TrimSpace(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func slugify(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return out
}

func DefaultServiceFromSource(source string) string {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		u, err := url.Parse(source)
		if err != nil {
			return ""
		}
		base := filepath.Base(u.Path)
		if base == "" || base == "/" {
			return ""
		}
		return slugify(strings.TrimSuffix(base, filepath.Ext(base)))
	}
	base := filepath.Base(source)
	return slugify(strings.TrimSuffix(base, filepath.Ext(base)))
}

func ResolveServiceName(serviceOverride, title, sourceHint string) (string, error) {
	if strings.TrimSpace(serviceOverride) != "" {
		s := sanitizeServiceSlug(serviceOverride)
		if s == "" {
			return "", fmt.Errorf("--service value is empty after sanitization")
		}
		return s, nil
	}
	if s := sanitizeServiceSlug(slugify(title)); s != "" {
		return s, nil
	}
	if s := sanitizeServiceSlug(DefaultServiceFromSource(sourceHint)); s != "" {
		return s, nil
	}
	return "", fmt.Errorf("cannot derive service: set --service or info.title, or use a spec filename")
}

func sanitizeServiceSlug(s string) string {
	out := slugify(strings.TrimSpace(s))
	if out == "" {
		return ""
	}
	r, _ := utf8.DecodeRuneInString(out)
	if r != utf8.RuneError && unicode.IsDigit(r) {
		return "s-" + out
	}
	return out
}

type NestedSpec struct {
	Title   string    `json:"title"`
	Version string    `json:"version"`
	Paths   *PathNode `json:"paths"`
}

type PathNode struct {
	Operation *OperationSpec      `json:"operation,omitempty"`
	Children  map[string]*PathNode `json:"children,omitempty"`
}

type OperationSpec struct {
	ID           string          `json:"id"`
	PathTemplate string          `json:"pathTemplate"`
	PathFormat   string          `json:"pathFormat"`
	PathArgNames []string        `json:"pathArgNames"`
	QueryParams  []ParameterSpec `json:"queryParams"`
	HeaderParams []ParameterSpec `json:"headerParams"`
}

type ParameterSpec struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

func (n *PathNode) insert(segments []string, op *OperationSpec) error {
	if n.Children == nil {
		n.Children = map[string]*PathNode{}
	}
	if len(segments) == 0 {
		if n.Operation != nil {
			return fmt.Errorf("duplicate GET for path %s", op.PathTemplate)
		}
		if len(n.Children) > 0 {
			if _, exists := n.Children["_"]; exists {
				return fmt.Errorf("ambiguous path tree at %s", op.PathTemplate)
			}
			n.Children["_"] = &PathNode{Operation: op}
			return nil
		}
		n.Operation = op
		return nil
	}
	key := segments[0]
	if n.Operation != nil {
		if _, exists := n.Children["_"]; exists {
			return fmt.Errorf("ambiguous path tree")
		}
		n.Children["_"] = &PathNode{Operation: n.Operation}
		n.Operation = nil
	}
	child := n.Children[key]
	if child == nil {
		child = &PathNode{}
		n.Children[key] = child
	}
	return child.insert(segments[1:], op)
}

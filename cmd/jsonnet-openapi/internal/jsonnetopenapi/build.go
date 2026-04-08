package jsonnetopenapi

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

func BuildPayload(api APISpec, serviceOverride string, sourceHint string) (*GenPayload, error) {
	api.GETOperations = normalizeGETOperations(api.GETOperations)
	title := api.Title
	version := api.Version
	root := &TrieNode{}
	count := 0
	for _, op := range api.GETOperations {
		geo, err := operationFromGET(op)
		if err != nil {
			return nil, err
		}
		segs := pathSegments(op.Path)
		err = root.insert(segs, geo)
		if err != nil {
			return nil, err
		}
		count++
	}
	if count == 0 {
		return nil, fmt.Errorf("no GET operations in spec")
	}
	svc, err := resolveServiceName(serviceOverride, title, sourceHint)
	if err != nil {
		return nil, err
	}
	return &GenPayload{
		Info: GenInfo{
			Title:   title,
			Version: version,
		},
		Service: svc,
		Trie:    root,
	}, nil
}

func normalizeGETOperations(ops []GETOperation) []GETOperation {
	out := make([]GETOperation, len(ops))
	for i := range ops {
		out[i] = ops[i]
		out[i].Parameters = canonicalizeParameters(ops[i].Parameters)
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

func operationFromGET(op GETOperation) (*GenOperation, error) {
	id := op.OperationID
	if id == "" {
		id = "get_" + slugify(strings.Trim(op.Path, "/"))
	}
	pf, pargs := pathFormatFromTemplate(op.Path)
	qp, hp := splitParams(op.Parameters)
	return &GenOperation{
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

func splitParams(params []Parameter) (query []ParamSpec, header []ParamSpec) {
	query = []ParamSpec{}
	header = []ParamSpec{}
	for _, p := range params {
		ps := ParamSpec{Name: p.Name, Required: p.Required}
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

func defaultServiceFromSource(source string) string {
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

func resolveServiceName(serviceOverride, title, sourceHint string) (string, error) {
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
	if s := sanitizeServiceSlug(defaultServiceFromSource(sourceHint)); s != "" {
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

func (n *TrieNode) insert(segments []string, op *GenOperation) error {
	if n.Children == nil {
		n.Children = map[string]*TrieNode{}
	}
	if len(segments) == 0 {
		if n.Leaf != nil {
			return fmt.Errorf("duplicate GET for path %s", op.PathTemplate)
		}
		if len(n.Children) > 0 {
			if _, exists := n.Children["_"]; exists {
				return fmt.Errorf("ambiguous trie at %s", op.PathTemplate)
			}
			n.Children["_"] = &TrieNode{Leaf: op}
			return nil
		}
		n.Leaf = op
		return nil
	}
	key := segments[0]
	if n.Leaf != nil {
		if _, exists := n.Children["_"]; exists {
			return fmt.Errorf("ambiguous trie")
		}
		n.Children["_"] = &TrieNode{Leaf: n.Leaf}
		n.Leaf = nil
	}
	child := n.Children[key]
	if child == nil {
		child = &TrieNode{}
		n.Children[key] = child
	}
	return child.insert(segments[1:], op)
}

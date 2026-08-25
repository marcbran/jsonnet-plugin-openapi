package links

import "encoding/json"

func overrideLinksSourcePath(input, output []byte) ([]byte, error) {
	var in struct {
		SourcePath json.RawMessage `json:"sourcePath"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, err
	}
	if in.SourcePath == nil {
		return output, nil
	}
	var out struct {
		Links []map[string]json.RawMessage `json:"links"`
	}
	if err := json.Unmarshal(output, &out); err != nil {
		return nil, err
	}
	for _, link := range out.Links {
		link["sourcePath"] = in.SourcePath
	}
	return json.Marshal(out)
}

func overrideEchoedFields(input, output []byte, fields ...string) ([]byte, error) {
	var in map[string]json.RawMessage
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, err
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(output, &out); err != nil {
		return nil, err
	}
	for _, field := range fields {
		if v, ok := in[field]; ok {
			out[field] = v
		}
	}
	return json.Marshal(out)
}

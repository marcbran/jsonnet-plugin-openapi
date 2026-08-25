package links

import "testing"

func TestOverrideLinksSourcePath(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		want   string
	}{
		{
			name:   "replaces empty echo",
			input:  `{"sourcePath":"/users/{username}","detailPaths":["/users/{username}"]}`,
			output: `{"links":[{"sourcePath":"","at":["owner"],"targetPath":"/users/{username}","keys":[{"const":"owner","path":null}],"confidence":"high","reason":"r"}]}`,
			want:   `{"links":[{"at":["owner"],"confidence":"high","keys":[{"const":"owner","path":null}],"reason":"r","sourcePath":"/users/{username}","targetPath":"/users/{username}"}]}`,
		},
		{
			name:   "replaces wrong echo",
			input:  `{"sourcePath":"/orgs/{org}/packages/{package_type}/{package_name}"}`,
			output: `{"links":[{"sourcePath":"/orgs/{org}","at":[],"targetPath":"/orgs/{org}","keys":[],"confidence":"low","reason":"r"}]}`,
			want:   `{"links":[{"at":[],"confidence":"low","keys":[],"reason":"r","sourcePath":"/orgs/{org}/packages/{package_type}/{package_name}","targetPath":"/orgs/{org}"}]}`,
		},
		{
			name:   "no links",
			input:  `{"sourcePath":"/users/{username}"}`,
			output: `{"links":[]}`,
			want:   `{"links":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := overrideLinksSourcePath([]byte(tt.input), []byte(tt.output))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestOverrideEchoedFields(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		fields []string
		want   string
	}{
		{
			name:   "replaces identity fields",
			input:  `{"sourcePath":"/accounts/{id}","targetPath":"/users/{userId}","at":["members"],"keys":[{"const":"members","path":null}],"sourceParams":["id"],"itemSchema":{}}`,
			output: `{"sourcePath":"","targetPath":"/users/{userId}","at":["members"],"keys":[{"const":"members","path":null}],"vars":[{"param":"userId","path":["userId"]}],"confidence":"high","reason":"r"}`,
			fields: []string{"sourcePath", "targetPath", "at", "keys"},
			want:   `{"at":["members"],"confidence":"high","keys":[{"const":"members","path":null}],"reason":"r","sourcePath":"/accounts/{id}","targetPath":"/users/{userId}","vars":[{"param":"userId","path":["userId"]}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := overrideEchoedFields([]byte(tt.input), []byte(tt.output), tt.fields...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

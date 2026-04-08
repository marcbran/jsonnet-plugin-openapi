package main

import (
	"os"
	"strings"

	"github.com/marcbran/jsonnet-plugin-openapi/openapi"
)

func main() {
	name := os.Getenv("OPENAPI_PLUGIN_NAME")
	if strings.TrimSpace(name) == "" {
		name = "openapi"
	}
	var opts []openapi.Option
	base := os.Getenv("OPENAPI_BASE_URL")
	if strings.TrimSpace(base) != "" {
		opts = append(opts, openapi.WithBaseURL(base))
	}
	openapi.NewPlugin(name, opts...).Serve()
}

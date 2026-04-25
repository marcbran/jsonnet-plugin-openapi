//go:build e2e

package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-jsonnet/formatter"
	"github.com/marcbran/jpoet/pkg/jpoet"
	pluginhttp "github.com/marcbran/jsonnet-plugin-http/http"
	openapipkg "github.com/marcbran/jsonnet-plugin-openapi/cmd/jsonnet-openapi/pkg/jsonnetopenapi"
	"github.com/stretchr/testify/require"
)

func testdataRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata")
}

func (s *Stage) a_spec(name string) *Stage {
	s.ref = filepath.Join(testdataRoot(), name+".yaml")
	return s
}

func (s *Stage) a_service(name string) *Stage {
	s.service = name
	return s
}

func (s *Stage) a_docker_container(image string, internalPort int) *Stage {
	t := s.t.(*testing.T)
	_, err := exec.LookPath("docker")
	if err != nil {
		t.Skip("docker not available in PATH")
	}
	runCmd := exec.Command(
		"docker",
		"run",
		"-d",
		"--rm",
		"-p",
		fmt.Sprintf("0:%d", internalPort),
		image,
	)
	var runStdout, runStderr bytes.Buffer
	runCmd.Stdout = &runStdout
	runCmd.Stderr = &runStderr
	err = runCmd.Run()
	require.NoError(t, err, "docker run: stdout=%q stderr=%q", runStdout.String(), runStderr.String())
	containerID := strings.TrimSpace(runStdout.String())
	t.Cleanup(func() {
		killOut, killErr := exec.Command("docker", "kill", containerID).CombinedOutput()
		if killErr != nil {
			t.Logf("docker kill %s: %v: %s", containerID, killErr, string(killOut))
		}
	})
	portCmd := exec.Command("docker", "port", containerID, fmt.Sprintf("%d/tcp", internalPort))
	portOut, err := portCmd.CombinedOutput()
	require.NoError(t, err, "docker port %s %d/tcp: %s", containerID, internalPort, string(portOut))
	portLine := strings.TrimSpace(string(portOut))
	hostPort := portLine[strings.LastIndex(portLine, ":")+1:]
	s.liveHTTPOrigin = "http://127.0.0.1:" + hostPort
	readyURL := s.liveHTTPOrigin + "/-/ready"
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(2 * time.Minute)
	var lastReadyErr error
	var lastReadyCode int
	for time.Now().Before(deadline) {
		resp, err := client.Get(readyURL)
		if err != nil {
			lastReadyErr = err
			time.Sleep(200 * time.Millisecond)
			continue
		}
		lastReadyCode = resp.StatusCode
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return s
		}
		time.Sleep(200 * time.Millisecond)
	}
	psOut, _ := exec.Command("docker", "ps", "-a", "--filter", "id="+containerID, "--format", "{{.Status}}").CombinedOutput()
	t.Fatalf(
		"docker container %s not ready: GET %s last_err=%v last_code=%d docker_ps_filter=%q",
		containerID,
		readyURL,
		lastReadyErr,
		lastReadyCode,
		strings.TrimSpace(string(psOut)),
	)
	return s
}

func (s *Stage) a_spec_url(url string) *Stage {
	if s.liveHTTPOrigin != "" && strings.HasPrefix(url, "/") {
		s.ref = strings.TrimSuffix(s.liveHTTPOrigin, "/") + url
		return s
	}
	s.ref = url
	return s
}

func (s *Stage) the_gen_command_is_run() *Stage {
	out, err := s.facade.Generate(context.Background(), openapipkg.Input{
		Ref:     s.ref,
		OutDir:  s.outDir,
		Service: s.service,
		PkgRepo: "git@github.com:marcbran/jsonnet.git",
	})
	if err != nil {
		s.lastOutput = out
		s.lastErr = err.Error()
		return s
	}
	s.lastOutput = out
	s.lastErr = ""
	return s
}

func (s *Stage) the_gen_has_no_error() *Stage {
	require.Empty(s.t, s.lastErr)
	return s
}

func (s *Stage) the_generated_files_match(name string) *Stage {
	expectedDir := filepath.Join(testdataRoot(), name)
	names := []string{"main.libsonnet", "pkg.libsonnet"}
	for _, fname := range names {
		gotPath := filepath.Join(s.outDir, fname)
		wantPath := filepath.Join(expectedDir, fname)
		got, err := os.ReadFile(gotPath)
		require.NoError(s.t, err)
		want, err := os.ReadFile(wantPath)
		require.NoError(s.t, err)
		require.Equal(s.t, string(want), string(got))
	}
	return s
}

func (s *Stage) the_generated_main_libsonnet_parses_as_jsonnet() *Stage {
	src, err := os.ReadFile(filepath.Join(s.outDir, "main.libsonnet"))
	require.NoError(s.t, err)
	_, _, err = formatter.SnippetToRawAST("main.libsonnet", string(src))
	require.NoError(s.t, err)
	return s
}

func (s *Stage) a_jsonnet_request_is_evaluated(baseURL string, expr string) *Stage {
	if s.liveHTTPOrigin != "" && strings.HasPrefix(baseURL, "/") {
		baseURL = strings.TrimSuffix(s.liveHTTPOrigin, "/") + baseURL
	}
	snippet := fmt.Sprintf("local %s = import 'main.libsonnet';\n%s", s.service, expr)
	var out map[string]any
	s.evalErr = jpoet.Eval(
		jpoet.WithPlugin(pluginhttp.Plugin(s.service, pluginhttp.WithBaseURL(baseURL))),
		jpoet.FileImport([]string{s.outDir}),
		jpoet.SnippetInput("eval.jsonnet", snippet),
		jpoet.ValueOutput(&out),
		jpoet.Serialize(false),
	)
	s.evalOut = out
	return s
}

func (s *Stage) the_eval_has_no_error() *Stage {
	require.NoError(s.t, s.evalErr)
	return s
}

func (s *Stage) the_result_has_status(want string) *Stage {
	require.NoError(s.t, s.evalErr)
	got, ok := s.evalOut["status"].(string)
	require.True(s.t, ok)
	require.Equal(s.t, want, got)
	return s
}

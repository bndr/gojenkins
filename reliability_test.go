package gojenkins

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallPluginChecksErrorBeforeStatus(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "crumbIssuer") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("hijacking not supported")
		}
		conn, _, err := hj.Hijack()
		require.NoError(t, err)
		_ = conn.Close()
	}))
	defer server.Close()

	err := jenkins.InstallPlugin(context.Background(), "git", "4.0")
	require.Error(t, err)
}

func TestPipelineGetArtifactsUnmarshalsSlice(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasPrefix(r.URL.Path, "/job/demo/1/wfapi/artifacts"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]PipelineArtifact{
			{ID: "a1", Name: "out.txt", Path: "out.txt", URL: "/artifact/out.txt", Size: 12},
			{ID: "a2", Name: "log.txt", Path: "log.txt", URL: "/artifact/log.txt", Size: 34},
		})
	}))
	defer server.Close()

	pr := &PipelineRun{
		Job:  &Job{Jenkins: jenkins, Base: "/job/demo"},
		Base: "/job/demo/1",
	}
	artifacts, err := pr.GetArtifacts(context.Background())
	require.NoError(t, err)
	require.Len(t, artifacts, 2)
	assert.Equal(t, "out.txt", artifacts[0].Name)
	assert.Equal(t, "log.txt", artifacts[1].Name)
}

func TestGetBuildFromQueueIDRespectsCanceledContext(t *testing.T) {
	polls := 0
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/queue/item/") {
			polls++
			_ = json.NewEncoder(w).Encode(taskResponse{
				ID: 7,
				Executable: struct {
					Number int64  `json:"number"`
					URL    string `json:"url"`
				}{Number: 0},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	job := &Job{Jenkins: jenkins, Raw: new(JobResponse), Base: "/job/demo"}
	_, err := jenkins.GetBuildFromQueueID(ctx, job, 7)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
	assert.GreaterOrEqual(t, polls, 1)
}

func TestArtifactSaveToDirPropagatesSaveError(t *testing.T) {
	dir := t.TempDir()
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	a := Artifact{
		Jenkins:  jenkins,
		FileName: "missing.bin",
		Path:     "/missing",
	}
	saved, err := a.SaveToDir(context.Background(), dir)
	assert.False(t, saved)
	require.Error(t, err)
	_, statErr := os.Stat(filepath.Join(dir, a.FileName))
	assert.True(t, os.IsNotExist(statErr))
}

func TestProceedInputRequiresPendingActions(t *testing.T) {
	jenkins, server := newTestJenkins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	pr := &PipelineRun{
		Job:  &Job{Jenkins: jenkins, Base: "/job/demo"},
		Base: "/job/demo/1",
	}
	ok, err := pr.ProceedInput(context.Background())
	assert.False(t, ok)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no pending input")
}

func TestJobParentBaseMissingSegment(t *testing.T) {
	j := &Job{Base: "/computer/agent"}
	assert.Equal(t, "", j.parentBase())
}

func TestFolderParentBaseMissingSegment(t *testing.T) {
	f := &Folder{Base: "/view/all"}
	assert.Equal(t, "", f.parentBase())
}

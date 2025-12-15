package gojenkins

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestArtifactGetDataSuccess tests successful artifact data retrieval
func TestArtifactGetDataSuccess(t *testing.T) {
	expectedData := "binary file content here"

	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
		GetFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if strPtr, ok := response.(*string); ok {
				*strPtr = expectedData
			}
			return &http.Response{StatusCode: http.StatusOK}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	artifact := Artifact{
		Jenkins:  jenkins,
		FileName: "app.jar",
		Path:     "/job/TestJob/123/artifact/target/app.jar",
	}

	ctx := context.Background()
	data, err := artifact.GetData(ctx)

	assert.NoError(t, err)
	assert.Equal(t, []byte(expectedData), data)
}

// TestArtifactGetDataNotFound tests artifact not found
func TestArtifactGetDataNotFound(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusNotFound},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	artifact := Artifact{
		Jenkins:  jenkins,
		FileName: "app.jar",
		Path:     "/job/TestJob/123/artifact/target/app.jar",
	}

	ctx := context.Background()
	data, err := artifact.GetData(ctx)

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "could not get File Contents")
}

// TestArtifactGetDataError tests artifact retrieval with network error
func TestArtifactGetDataError(t *testing.T) {
	mock := &MockRequester{
		err: assert.AnError,
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	artifact := Artifact{
		Jenkins:  jenkins,
		FileName: "app.jar",
		Path:     "/job/TestJob/123/artifact/target/app.jar",
	}

	ctx := context.Background()
	data, err := artifact.GetData(ctx)

	assert.Error(t, err)
	assert.Nil(t, data)
}

// TestArtifactGetMD5Local tests MD5 calculation for local files
func TestArtifactGetMD5Local(t *testing.T) {
	artifact := Artifact{
		FileName: "test.txt",
	}

	// Test with non-existent file - should return empty string
	hash := artifact.getMD5local("/nonexistent/path/file.txt")
	assert.Empty(t, hash)
}

// TestArtifactPathConstruction verifies artifact path is correctly used
func TestArtifactPathConstruction(t *testing.T) {
	var capturedEndpoint string

	mock := &MockRequester{
		response: &http.Response{StatusCode: http.StatusOK},
		GetFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			capturedEndpoint = endpoint
			if strPtr, ok := response.(*string); ok {
				*strPtr = "data"
			}
			return &http.Response{StatusCode: http.StatusOK}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	artifact := Artifact{
		Jenkins:  jenkins,
		FileName: "output.zip",
		Path:     "/job/MyJob/42/artifact/dist/output.zip",
	}

	ctx := context.Background()
	_, _ = artifact.GetData(ctx)

	assert.Equal(t, "/job/MyJob/42/artifact/dist/output.zip", capturedEndpoint)
}

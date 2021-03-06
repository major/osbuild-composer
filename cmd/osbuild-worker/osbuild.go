package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/osbuild/osbuild-composer/internal/common"
	"github.com/osbuild/osbuild-composer/internal/distro"
)

type OSBuildError struct {
	Message string
	Result  *common.ComposeResult
}

func (e *OSBuildError) Error() string {
	return e.Message
}

func RunOSBuild(manifest distro.Manifest, store, outputDirectory string, errorWriter io.Writer) (*common.ComposeResult, error) {
	cmd := exec.Command(
		"osbuild",
		"--store", store,
		"--output-directory", outputDirectory,
		"--json", "-",
	)
	cmd.Stderr = errorWriter

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("error setting up stdin for osbuild: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error setting up stdout for osbuild: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting osbuild: %v", err)
	}

	err = json.NewEncoder(stdin).Encode(manifest)
	if err != nil {
		return nil, fmt.Errorf("error encoding osbuild pipeline: %v", err)
	}
	// FIXME: handle or comment this possible error
	_ = stdin.Close()

	var result common.ComposeResult
	err = json.NewDecoder(stdout).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error decoding osbuild output: %#v", err)
	}

	err = cmd.Wait()
	if err != nil {
		return nil, &OSBuildError{
			Message: fmt.Sprintf("running osbuild failed: %v", err),
			Result:  &result,
		}
	}

	return &result, nil
}

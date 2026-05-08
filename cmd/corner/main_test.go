package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunReturnsErrorOnEmptyArgs(t *testing.T) {
	err := run(nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error for empty args")
	}
	if !strings.Contains(err.Error(), "argv[0]") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunMooPrintsASCII(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"corner", "--moo"}, &out, io.Discard)
	if err != nil {
		t.Fatalf("run moo: %v", err)
	}
	if !strings.Contains(out.String(), "(__)") {
		t.Fatalf("expected moo ascii output, got %q", out.String())
	}
}

func TestRunReturnsConfigError(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	err := run([]string{"corner", "--quality", "999"}, &out, &errOut)
	if err == nil {
		t.Fatal("expected config validation error")
	}
	if !strings.Contains(err.Error(), "quality") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReturnsPipelineError(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err = run([]string{"corner", "--mask", "*.jpg", "--out-dir", "out"}, &out, io.Discard)
	if err == nil {
		t.Fatal("expected pipeline error for empty input")
	}
	if !strings.Contains(err.Error(), "no files found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

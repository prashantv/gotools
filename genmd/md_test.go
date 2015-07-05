package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testFile(inFile string, expected string) error {
	file, err := os.Open(inFile)
	if err != nil {
		return err
	}

	expectedBytes, err := ioutil.ReadFile(expected)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err = processMarkdown(file, buf); err != nil {
		return err
	}

	if !bytes.Equal(expectedBytes, buf.Bytes()) {
		cmd := exec.Command("diff", expected, "-")
		cmd.Stdin = buf
		cmd.Stdout = os.Stdout
		cmd.Run()
		return fmt.Errorf("output mismatch (ran diff)")
	}

	return nil
}

func TestAllFiles(t *testing.T) {
	files, err := ioutil.ReadDir("test_files")
	if err != nil {
		t.Fatalf("Failed to read test_files directory: %v", err)
	}

	for _, f := range files {
		fname := f.Name()
		if !strings.HasSuffix(fname, ".md.in") {
			continue
		}

		fname = filepath.Join("test_files", fname)
		outFile := strings.TrimSuffix(fname, ".in")
		fmt.Printf("Test %v -> %v\n", fname, outFile)
		if err := testFile(fname, outFile); err != nil {
			t.Errorf("test %v -> %v failed: %v", fname, outFile, err)
		}
	}
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

// godocdown generates markdown with documentation and code
// but the code is not marked with the ```go tag.
// The output looks like this:
// another line of text
//     code1
//
//     code2
//     code3
//
// documentation resume
// more documentation

// This parser runs godocdown and detects code blocks, and replaces
// them with ```go code blocks.

const codePrefix = "    "

func main() {
	inFile := "README.md.tmpl"
	if len(os.Args) > 1 {
		inFile = os.Args[1]
	}

	md, err := runGoDocDown(inFile)
	if err != nil {
		log.Fatalf("failed to run godocdown: %v", err)
	}
	defer md.Close()
	if err := processMarkdown(md, os.Stdout); err != nil {
		log.Fatalf("process output failed: %v", err)
	}
}

func processMarkdown(md io.Reader, output io.Writer) error {
	const codeStart = "```go"
	const codeEnd = "```\n" // Extra new line after a code block.

	var inCodeBlock bool
	for rdr := bufio.NewReader(md); ; {
		line, err := rdr.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if !inCodeBlock {
			if strings.HasPrefix(line, codePrefix) {
				fmt.Fprintln(output, codeStart)
				inCodeBlock = true
			}
		}

		if inCodeBlock {
			if strings.HasPrefix(line, codePrefix) {
				line = strings.TrimPrefix(line, codePrefix)
			} else if line == "\n" {
				// Code blocks end with a blank line which we want to avoid.
				bs, err := rdr.Peek(1)
				if (err == nil && !unicode.IsSpace(rune(bs[0]))) ||
					err == io.EOF {
					line = ""
				}
			} else {
				fmt.Fprintln(output, codeEnd)
				inCodeBlock = false
			}
		}

		fmt.Fprint(output, line)
	}

	// If there's a code block at the end of the file, end it.
	if inCodeBlock {
		fmt.Fprint(output, codeEnd)
	}
	return nil
}

func runGoDocDown(tmplFile string) (io.ReadCloser, error) {
	cmd := exec.Command("godocdown", "-template", tmplFile)
	output, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		output.Close()
		return nil, err
	}
	return output, nil
}

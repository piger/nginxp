package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/piger/nginxp/internal/parse"
)

var (
	magicHeader    = regexp.MustCompile(`^# configuration file ([^:]+):\n$`)
	defaultSection = "__DEFAULT__"
)

var usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: <filename> [section]\n", os.Args[0])
	flag.PrintDefaults()
}

// unpackConfigurationDump reads a configuration dump produced by "nginx -T" and extracts each file
// separately; the returned map maps the name of the files in the configuration dump with their contents.
// This function also supports reading a single configuration file, in which case its contents will be stored
// in the special "__DEFAULT__" key.
func unpackConfigurationDump(filename string) (map[string]string, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	contentMap := make(map[string]string)
	current := defaultSection
	var lines []string

	r := bufio.NewReader(fh)

Loop:
	for {
		line, err := r.ReadString('\n')
		line = strings.ReplaceAll(line, "\r", "")
		switch {
		case errors.Is(err, io.EOF):
			break Loop
		case err != nil:
			return nil, err
		}

		m := magicHeader.FindStringSubmatch(line)
		switch {
		case m == nil:
			lines = append(lines, line)
		case len(m) == 1:
			return nil, fmt.Errorf("invalid magic header line: %q", line)
		case len(m) > 1:
			if len(lines) > 0 {
				contentMap[current] = strings.Join(lines, "")
			}
			lines = nil
			current = m[1]
		}
	}

	contentMap[current] = strings.Join(lines, "")

	return contentMap, nil
}

func run() error {
	filename := flag.Arg(0)
	if filename == "" {
		flag.Usage()
		return errors.New("missing filename parameter")
	}

	var section string
	if len(flag.Args()) > 1 {
		section = flag.Arg(1)
	}

	filesMap, err := unpackConfigurationDump(filename)
	if err != nil {
		return err
	}

	if section != "" {
		contents, ok := filesMap[section]
		if !ok {
			return fmt.Errorf("section %q not found", section)
		}
		parse.LexerPlayground(section, contents)
	} else {
		contents, ok := filesMap[defaultSection]
		if !ok {
			return errors.New("a configuration dump was read but no section was specified")
		}
		parse.LexerPlayground(defaultSection, contents)
	}

	return nil
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if err := run(); err != nil {
		fmt.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

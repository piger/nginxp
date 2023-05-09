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

	flagTestLexer  = flag.Bool("lex", false, "Show the output of the lexer")
	flagAllSection = flag.Bool("all", false, "Parse all sections in a configuration dump")
	flagPlayground = flag.Bool("play", false, "Call the playground function")
	flagStuff      = flag.Bool("stuff", false, "Run testing stuff")
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

	switch {
	case *flagAllSection:
		for name, contents := range filesMap {
			// for now this program doesn't support parsing included map files so we just print them verbatim.
			if strings.HasSuffix(name, ".map") {
				fmt.Print(contents)
			} else {
				return play(name, contents)
			}
		}
	case section != "":
		contents, ok := filesMap[section]
		if !ok {
			return fmt.Errorf("section %q not found", section)
		}
		return play(section, contents)
	default:
		contents, ok := filesMap[defaultSection]
		if !ok {
			fmt.Printf("Available sections:\n")
			for section, data := range filesMap {
				fmt.Printf("\t%s (%d bytes)\n", section, len(data))
			}
			return errors.New("a configuration dump was read but no section was specified")
		}
		return play(defaultSection, contents)
	}

	return nil
}

func play(filename string, contents string) error {
	if *flagPlayground {
		parse.LexerPlayground(filename, contents, *flagTestLexer)
		return nil
	} else {
		return parse.Analyse(filename, contents)
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if err := run(); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

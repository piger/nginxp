package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/piger/nginxp/internal/parse"
)

var (
	flagTestLexer  = flag.Bool("lex", false, "Show the output of the lexer")
	flagAllSection = flag.Bool("all", false, "Parse all sections in a configuration dump")
	flagPlayground = flag.Bool("play", false, "Call the playground function")
	flagStuff      = flag.Bool("stuff", false, "Run testing stuff")
)

var usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: <filename> [section]\n", os.Args[0])
	flag.PrintDefaults()
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

	filesMap, err := parse.Unpack(filename)
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
		contents, ok := filesMap[filename]
		if !ok {
			fmt.Printf("Available sections:\n")
			for section, data := range filesMap {
				fmt.Printf("\t%s (%d bytes)\n", section, len(data))
			}
			return errors.New("a configuration dump was read but no section was specified")
		}
		return play(filename, contents)
	}

	return nil
}

func play(filename string, contents string) error {
	if *flagPlayground {
		parse.LexerPlayground(filename, contents, *flagTestLexer)
		return nil
	} else {
		tree, err := parse.Parse(filename, contents)
		if err != nil {
			return err
		}

		cfg, err := parse.NewConfiguration(tree)
		if err != nil {
			return err
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(cfg); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if err := run(); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

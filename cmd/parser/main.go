package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/piger/nginxp/internal/parse"
)

func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func run() error {
	filename := flag.Arg(0)
	if filename == "" {
		return errors.New("missing filename parameter")
	}

	contents, err := readFile(filename)
	if err != nil {
		return err
	}

	parse.LexerPlayground(filename, contents)
	return nil
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}

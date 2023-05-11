package parse

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var magicHeader = regexp.MustCompile(`^# configuration file ([^:]+):\r?\n$`)

// Unpack reads a file containing the output of `nginx -T` and returns a map
// where the keys are the filenames found in the configuration file, and the keys
// are their contents.
//
// If filename contains a single nginx configuration file instead of a configuration dump,
// the only key in the returned map will be filename.
func Unpack(filename string) (map[string]string, error) {
	files := make(map[string]string)

	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	r := bufio.NewReader(fh)
	b := strings.Builder{}
	currentFilename := filename

Loop:
	for {
		line, err := r.ReadString('\n')
		switch {
		case errors.Is(err, io.EOF):
			break Loop
		case err != nil:
			return nil, err
		}

		line = strings.ReplaceAll(line, "\r", "")
		match := magicHeader.FindStringSubmatch(line)
		switch {
		case match == nil:
			b.WriteString(line)
		case len(match) == 1:
			return nil, fmt.Errorf("invalid magic header line: %q", line)
		case len(match) > 1:
			if b.Len() > 0 {
				files[currentFilename] = b.String()
			}
			b.Reset()
			currentFilename = string(match[1])
		}
	}

	// don't forget to add the last file! :)
	if b.Len() > 0 {
		files[currentFilename] = b.String()
	}

	return files, nil
}

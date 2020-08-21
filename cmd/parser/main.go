package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/piger/nginxp"
)

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		fmt.Println("missing filename")
		return
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	nginxp.Stuff(string(content))
}

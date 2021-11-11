package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/qawatake/obsdconv/process"
)

var (
	Version = "x.y.z"
)

const (
	DEFAULT_IGNORE_FILE_NAME = ".obsdconvignore"
)

func main() {
	var flags flagBundle
	initFlags(flag.CommandLine, &flags)
	flag.Parse()
	if flags.ver {
		fmt.Printf("v%s\n", Version)
		return
	}
	if err := setFlags(flag.CommandLine, &flags); err != nil {
		log.Fatal(err)
	}
	if err := verifyFlags(&flags); err != nil {
		log.Fatal(err)
	}
	processor := newDefaultProcessor(&flags)
	skipper, err := process.NewSkipper(DEFAULT_IGNORE_FILE_NAME)
	if err != nil {
		log.Fatal(err)
	}
	if err := process.Walk(flags.src, flags.dst, skipper, processor); err != nil {
		log.Fatal(err)
	}
}

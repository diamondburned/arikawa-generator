package main

import (
	"flag"
	"go/format"
	"io"
	"log"
	"os"
	"strings"

	_ "embed"

	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/pb33f/libopenapi"
)

//go:embed initials.txt
var initialsFile string

func init() {
	initials := strings.Fields(initialsFile)
	strcases.AddPascalSpecials(initials)
}

//go:embed openapi.json
var openapiJSON []byte

var (
	outputFile = "-"
	outputPkg  = "main"
)

func init() {
	flag.StringVar(&outputFile, "o", outputFile, "output file")
	flag.StringVar(&outputPkg, "pkg", outputPkg, "output package")
}

func main() {
	flag.Parse()

	doc, err := libopenapi.NewDocument(openapiJSON)
	if err != nil {
		log.Fatalln(err)
	}

	code, err := Generate(doc, outputPkg)
	if err != nil {
		log.Fatalln(err)
	}

	if formatted, err := format.Source(code); err == nil {
		code = formatted
	}

	var out io.WriteCloser
	if outputFile == "-" {
		out = os.Stdout
	} else {
		out, err := os.Create(outputFile)
		if err != nil {
			log.Fatalln(err)
		}
		defer out.Close()
	}

	if _, err := out.Write(code); err != nil {
		log.Fatalln(err)
	}

	if err := out.Close(); err != nil {
		log.Fatalln(err)
	}
}

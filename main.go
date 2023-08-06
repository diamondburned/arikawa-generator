package main

import (
	"flag"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/hashicorp/go-hclog"
	"github.com/pb33f/libopenapi"
)

var (
	outputFile          = "-"
	outputPkg           = "main"
	openapiFile         = filepath.Join(os.Getenv("DISCORD_API_SPEC"), "specs", "openapi.json")
	documentationDir    = filepath.Join(os.Getenv("DISCORD_API_DOCS"), "docs", "resources")
	initialsFile        string
	snowflakeFieldsFile string
)

func init() {
	hclog.Default().SetLevel(hclog.Debug)

	flag.StringVar(&outputFile, "o", outputFile, "output file")
	flag.StringVar(&outputPkg, "pkg", outputPkg, "output package")
	flag.StringVar(&openapiFile, "openapi", openapiFile, "openapi file")
	flag.StringVar(&documentationDir, "docs", documentationDir, "documentation directory")
	flag.StringVar(&initialsFile, "initials", initialsFile, "initials file")
	flag.StringVar(&snowflakeFieldsFile, "snowflake-fields", snowflakeFieldsFile, "snowflake fields file")
}

func main() {
	flag.Parse()

	if initialsFile != "" {
		b, err := os.ReadFile(initialsFile)
		if err != nil {
			log.Fatalln(err)
		}
		strcases.AddPascalSpecials(strings.Fields(string(b)))
	}

	if snowflakeFieldsFile != "" {
		b, err := os.ReadFile(snowflakeFieldsFile)
		if err != nil {
			log.Fatalln(err)
		}
		addSnowflakeFieldsFile(string(b))
	}

	if err := scrapeDocs(); err != nil {
		log.Fatalln(err)
	}

	openapiJSON, err := os.ReadFile(openapiFile)
	if err != nil {
		log.Fatalln(err)
	}

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
	} else {
		log.Println("cannot format code:", err)
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

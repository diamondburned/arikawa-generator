package main

import (
	"log"
	"strings"

	_ "embed"

	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

//go:embed data/initials.txt
var embeddedInitialsFile string

func init() {
	initials := strings.Fields(embeddedInitialsFile)
	strcases.AddPascalSpecials(initials)
}

//go:embed data/snowflake-fields.txt
var embeddedSnowflakeFields string

func init() {
	addSnowflakeFieldsFile(embeddedSnowflakeFields)
}

var snowflakeFields = map[string]string{}

func addSnowflakeFieldsFile(file string) {
	for _, line := range strings.Split(file, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		values := strings.Fields(line)
		if len(values) != 2 {
			log.Println("invalid snowflake field:", line)
			continue
		}
		snowflakeFields[values[0]] = values[1]
	}
}

//go:embed data/snowflakes.txt
var embeddedSnowflakes string

func init() {
	addSnowflakeFile(embeddedSnowflakes)
}

var snowflakes = NewSet[string]()

func addSnowflakeFile(file string) {
	for _, line := range strings.Split(file, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		snowflakes.Add(line)
	}
}

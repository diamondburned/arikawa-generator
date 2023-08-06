package docread

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"
)

// ObjectTable is a table of fields in an object.
type ObjectTable struct {
	Sections    []string /// path to the section
	Description string
	Table       []ObjectTableRow

	Source Source
}

func (t ObjectTable) FieldInfos() []FieldInfo {
	fields := make([]FieldInfo, 0, len(t.Table))
	for _, row := range t.Table {
		fields = append(fields, row.FieldInfo())
	}
	return fields
}

func (t ObjectTable) FieldTable() FieldTable {
	return FieldTable{
		Name:   strings.Join(t.Sections, " "),
		Fields: t.FieldInfos(),
	}
}

// ObjectTableRow is a row in an object table.
type ObjectTableRow struct {
	Field       string
	Type        string
	Description string
}

func (r ObjectTableRow) FieldInfo() FieldInfo {
	var f FieldInfo
	f.Name = strings.TrimSuffix(r.Field, "?")
	f.Type = strings.TrimPrefix(r.Type, "?")
	f.Comment = r.Description
	f.Optional = strings.HasSuffix(f.Name, "?")
	f.Nullable = strings.HasPrefix(f.Type, "?")
	return f
}

// Source describes where a table came from.
type Source struct {
	Path     string
	Position int
}

// ScrapeFS scrapes the documentation filesystem for object tables.
func ScrapeFS(dir fs.FS) ([]ObjectTable, error) {
	var tables []ObjectTable
	err := fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		b, err := fs.ReadFile(dir, path)
		if err != nil {
			return err
		}

		t, err := scrapeBytes(path, b)
		if err != nil {
			return err
		}
		tables = append(tables, t...)
		return nil
	})
	return tables, err
}

// ScrapeDir scrapes the given documentation directory for object tables.
func ScrapeDir(path string) ([]ObjectTable, error) {
	return ScrapeFS(os.DirFS(path))
}

var tableHeaderRe = regexp.MustCompile(`(?m)^\| Field +\| Type +\| Description +\|$`)

// ScrapeBytes scrapes the given documentation file's content for object tables.
func ScrapeBytes(b []byte) ([]ObjectTable, error) {
	return scrapeBytes("", b)
}

func scrapeBytes(path string, b []byte) ([]ObjectTable, error) {
	var tables []ObjectTable
	for _, pos := range tableHeaderRe.FindAllIndex(b, -1) {
		t, err := readTable(b, pos[0])
		if err != nil {
			return nil, fmt.Errorf("cannot read table at %d: %w", pos[0], err)
		}
		tables = append(tables, *t)
	}
	for i := range tables {
		tables[i].Source.Path = path
	}
	return tables, nil
}

var (
	rowRe    = regexp.MustCompile(`(?m)^\| (.*?) *\| (.*?) *\| (.*?) *\|$`)
	rowSepRe = regexp.MustCompile(`(?m)^\| -* \| -* \| -* \|$`)
	headerRe = regexp.MustCompile(`(?m)^(#{1,6}) +(.*)$`)
)

func readTable(src []byte, pos int) (*ObjectTable, error) {
	// Scan until we cannot match a table row.
	scanner := bufio.NewScanner(bytes.NewReader(src[pos:]))
	var tableRows []ObjectTableRow
	var scanned int
	for scanner.Scan() {
		line := scanner.Text()
		cols := rowRe.FindStringSubmatch(line)
		if cols == nil {
			break
		}

		if scanned < 2 {
			// Skip the header and the separator.
			scanned++
			continue
		}

		for i, col := range cols {
			cols[i] = strings.TrimSpace(col)
			cols[i] = strings.TrimSuffix(cols[i], " *")
		}

		tableRows = append(tableRows, ObjectTableRow{
			Field:       cols[1],
			Type:        cols[2],
			Description: cols[3],
		})
	}

	// Find the last # symbol up to pos. This is the current section.
	headers := headerRe.FindAllStringSubmatchIndex(string(src[:pos]), -1)
	if len(headers) == 0 {
		return nil, fmt.Errorf("table at %d has no section", pos)
	}

	var sections []string
	var lastHashes int
	for i := len(headers) - 1; i >= 0; i-- {
		hashes := bytes.Count(src[headers[i][2]:headers[i][3]], []byte{'#'})
		if lastHashes > 0 && hashes >= lastHashes {
			// This section is a subsection of some other section that we don't
			// care about.
			break
		}
		content := string(src[headers[i][4]:headers[i][5]])
		sections = append([]string{content}, sections...) // prepend
		lastHashes = hashes
	}

	// The description is everything between the last header and the table.
	lastSectionRange := headers[len(headers)-1]
	description := strings.TrimSpace(string(src[lastSectionRange[1]:pos]))

	return &ObjectTable{
		Sections:    sections, // TODO: implement
		Description: description,
		Table:       tableRows,
		Source: Source{
			Path:     "TODO",
			Position: pos,
		},
	}, nil
}

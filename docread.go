package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"sync"

	openapibase "github.com/pb33f/libopenapi/datamodel/high/base"
	"golang.org/x/exp/slices"
	"libdb.so/arikawa-generator/internal/docread"
)

var knownDocTables []docread.ObjectTable

func scrapeDocs() error {
	t, err := docread.ScrapeDir(documentationDir)
	if err != nil {
		return err
	}
	knownDocTables = append(knownDocTables, t...)
	return nil
}

const maxLikelihoodCandidates = 5
const minLikelihood = 5

var computedLikelihoods sync.Map // map[[32]byte][]docLikelihood

type docLikelihood struct {
	docread.ObjectTable
	Likelihood docread.ObjectLikelihood
}

func logLikelihoods(path schemaPath, candidates []docLikelihood) {
	var b strings.Builder
	fmt.Fprintln(&b, path, "might correlate to these tables:")
	for _, doc := range candidates {
		fmt.Fprintf(&b, "  - %.02f: %q\n", doc.Likelihood.Score, doc.Sections)
	}
	log.Println(b.String())
}

// calculateTopLikelihood returns the top 5 most likely objects that the given
// object is. The first object is the most likely, and the last object is the
// least likely.
func calculateTopLikelihood(path schemaPath, object *openapibase.Schema) []docLikelihood {
	id := object.GoLow().Hash()
	if cached, ok := computedLikelihoods.Load(id); ok {
		return cached.([]docLikelihood)
	}

	fields := objectToDocTable(path, object)
	candidates := make([]docLikelihood, 0, maxLikelihoodCandidates)

	for _, table := range knownDocTables {
		likelihood := docread.CalculateObjectLikelihood(fields, table.FieldTable())
		if likelihood.Score < minLikelihood {
			continue
		}

		if len(candidates) < maxLikelihoodCandidates {
			candidates = append(candidates, docLikelihood{
				ObjectTable: table,
				Likelihood:  likelihood,
			})
			continue
		}

		minIx := minLikelihoodCandidate(candidates)
		if likelihood.Score > candidates[minIx].Likelihood.Score {
			candidates[minIx] = docLikelihood{
				ObjectTable: table,
				Likelihood:  likelihood,
			}
		}
	}

	// Sort candidates by likelihood.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Likelihood.Score > candidates[j].Likelihood.Score
	})

	computedLikelihoods.Store(id, candidates)
	return candidates
}

func minLikelihoodCandidate(candidates []docLikelihood) int {
	min := 0
	for i, candidate := range candidates {
		if candidate.Likelihood.Score < candidates[min].Likelihood.Score {
			min = i
		}
	}
	return min
}

var capitalLetterRe = regexp.MustCompile(`[A-Z]`)

func objectToDocTable(path schemaPath, object *openapibase.Schema) docread.FieldTable {
	fields := make([]docread.FieldInfo, 0, len(object.Properties))
	for name, prop := range object.Properties {
		prop := prop.Schema()

		ptype, _ := extractPrimaryType(prop.Type)
		required := slices.Contains(object.Required, name)

		fields = append(fields, docread.FieldInfo{
			Name:     name,
			Type:     ptype.Type,
			Nullable: ptype.Nullable,
			Optional: !required,
		})
	}

	name := path.String()
	name = strings.ReplaceAll(name, ".", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = capitalLetterRe.ReplaceAllStringFunc(name, func(s string) string {
		return " " + strings.ToLower(s)
	})
	name = strings.TrimSpace(name)

	return docread.FieldTable{
		Name:   name,
		Fields: fields,
	}
}

func propIsNullable(object *openapibase.Schema, propName string) bool {
	required := slices.Contains(object.Required, propName)
	property := object.Properties[propName].Schema()
	ptype, _ := extractPrimaryType(property.Type)
	return !required || ptype.Nullable
}

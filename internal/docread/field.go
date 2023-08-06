package docread

import (
	"strings"

	"github.com/agnivade/levenshtein"
	"golang.org/x/exp/constraints"
)

// FieldTable is a table of fields. It is converted from an ObjectTable.
type FieldTable struct {
	Name   string
	Fields []FieldInfo
}

// FieldInfo is information about a field. It is converted from an
// ObjectTableRow.
type FieldInfo struct {
	Name     string
	Type     string
	Comment  string
	Optional bool
	Nullable bool
}

// ToFieldMap converts a slice of FieldInfo to a map of FieldInfo keyed by
// field name.
func ToFieldMap(infos []FieldInfo) map[string]FieldInfo {
	m := make(map[string]FieldInfo)
	for _, info := range infos {
		m[info.Name] = info
	}
	return m
}

// CalculateFieldLikelihood calculates the likelihood that a field is the same
// as another field. It returns a number at least 0; the higher the number, the
// more likely the fields are the same.
//
// Note that comments are ignored. Them being missing is the whole point of
// this function.
func CalculateFieldLikelihood(a, b FieldInfo) float64 {
	if a.Name != b.Name {
		return 0
	}
	return 0.5 + (relativeLevenshtein(a.Type, b.Type) / 2)
}

// ObjectLikelihood is the likelihood that an object is the same as another
// object. It is calculated from the likelihood of its fields.
type ObjectLikelihood struct {
	Score float64
}

// CalculateObjectLikelihood calculates the likelihood that an object is the
// same as another object. It returns a number at least 0; the higher the
// number, the more likely the objects are the same.
func CalculateObjectLikelihood(a, b FieldTable) ObjectLikelihood {
	const (
		nameWeight   = 4.0
		fieldsWeight = 9.0
	)
	likelihood := relativeLevenshtein(a.Name, b.Name) * nameWeight

	fieldTotal := min(len(a.Fields), len(b.Fields))
	if fieldTotal == 0 {
		return ObjectLikelihood{Score: likelihood}
	}

	bFields := ToFieldMap(b.Fields)
	fieldCount := 0
	fieldLikelihood := 0.0
	for _, aRow := range a.Fields {
		bRow, ok := bFields[aRow.Name]
		if !ok {
			continue
		}

		fieldCount++
		fieldLikelihood += CalculateFieldLikelihood(aRow, bRow)
	}

	if fieldCount > 0 {
		likelihood += fieldLikelihood / float64(fieldTotal) * fieldsWeight
	}

	return ObjectLikelihood{
		Score: likelihood,
	}
}

func relativeLevenshtein(a, b string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	maxDistance := float64(max(len(a), len(b)))
	if maxDistance == 0 {
		return 1
	}
	return 1 - (float64(levenshtein.ComputeDistance(a, b)) / maxDistance)
}

func min[T constraints.Float | constraints.Integer](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T constraints.Float | constraints.Integer](a, b T) T {
	if a > b {
		return a
	}
	return b
}

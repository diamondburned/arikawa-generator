package cmt

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type Opts struct {
	OriginalName string
}

func Prettify(self, cmt string, opts Opts) string {
	switch strings.ToLower(nthWord(cmt, 0)) {
	case "a", "an", "the":
		cmt = popFirstWord(cmt)
	}

	typeNamed := opts.OriginalName != "" &&
		strings.HasSuffix(nthWord(cmt, 0), opts.OriginalName)

	switch {
	case strings.ToLower(nthWord(cmt, 0)) == "is":
		fallthrough
	case strings.ToLower(nthWord(cmt, 0)) == "will":
		cmt = self + " " + lowerFirstWord(cmt)
	case strings.ToLower(nthWord(cmt, 0)) == "emitted":
		cmt = self + " is " + lowerFirstWord(cmt)
	case typeNamed:
		fallthrough
	case strings.HasPrefix(cmt, "#") && nthWord(cmt, 1) != "":
		cmt = self + " is the " + lowerFirstWord(cmt)
		// Trim the first word away and replace it with the Go name.
		// cmt = self + " " + popFirstWord(cmt)
	case nthWordSimplePresent(cmt, 0):
		cmt = self + " " + lowerFirstWord(cmt)
	default:
		// Trim the word "this" away to make the sentence gramatically
		// correct.
		cmt = strings.TrimPrefix(cmt, "this ")
		cmt = self + ": " + lowerFirstLetter(cmt)
	}

	cmt = addPeriod(cmt)

	return cmt
}

// popFirstWord pops the first word off.
func popFirstWord(paragraph string) string {
	parts := strings.SplitN(paragraph, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// lowerFirstLetter lower-cases the first letter in the paragraph.
func lowerFirstWord(paragraph string) string {
	r, sz := utf8.DecodeRuneInString(paragraph)
	if sz > 0 {
		return string(unicode.ToLower(r)) + paragraph[sz:]
	}
	return string(paragraph)
}

// nthWord returns the nth word, or an empty string if none.
func nthWord(paragraph string, n int) string {
	words := strings.SplitN(paragraph, " ", n+2)
	if len(words) < n+2 {
		return ""
	}
	return words[n]
}

// nthWordSimplePresent checks if the second word has a trailing "s".
func nthWordSimplePresent(paragraph string, n int) bool {
	word := nthWord(paragraph, n)
	return !strings.EqualFold(word, "this") && strings.HasSuffix(word, "s")
}

func lowerFirstLetter(p string) string {
	if p == "" {
		return ""
	}

	runes := []rune(p)
	if len(runes) < 2 {
		return string(unicode.ToLower(runes[0]))
	}

	// Edge case: gTK, etc.
	if unicode.IsUpper(runes[1]) {
		return p
	}

	return string(unicode.ToLower(runes[0])) + string(runes[1:])
}

func addPeriod(cmt string) string {
	if cmt != "" && !strings.HasSuffix(cmt, ".") {
		cmt += "."
	}
	return cmt
}

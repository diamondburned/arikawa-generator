package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	stderrors "errors"
	stdpath "path"

	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/pb33f/libopenapi"
	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"

	openapibase "github.com/pb33f/libopenapi/datamodel/high/base"
)

func init() {
	strcases.AddPascalSpecials([]string{
		"Afk",
		"Mfa",
		"Nsfw",
		"Rsvp",
	})
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

	var buf bytes.Buffer
	g := NewGenerator(&buf)
	if err := g.Generate(doc); err != nil {
		log.Fatalln(err)
	}

	if formatted, err := format.Source(buf.Bytes()); err == nil {
		buf.Reset()
		buf.Write(formatted)
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

	if _, err := buf.WriteTo(out); err != nil {
		log.Fatalln(err)
	}

	if err := out.Close(); err != nil {
		log.Fatalln(err)
	}
}

// Generator generates the code.
type Generator struct {
	output *bufio.Writer
	state  *generateState
}

type generateState struct {
	global *Generator
	errors generatorErrors
}

type generatorErrors struct {
	sync.Mutex
	errors []error
	count  int
}

func (e *generatorErrors) add(err error) {
	e.Lock()
	defer e.Unlock()

	e.count++
	if len(e.errors) < 10 {
		e.errors = append(e.errors, err)
	}
}

// NewGenerator returns a new generator.
func NewGenerator(w io.Writer) *Generator {
	state := &generateState{}
	global := &Generator{output: bufio.NewWriter(w), state: state}
	state.global = global
	return global
}

func (g *Generator) error(err error) {
	if err != nil {
		g.state.errors.add(err)
	}
}

func (g *Generator) skip(what string, why ...any) {
	if what != "" {
		what = strconv.Quote(what)
	} else {
		what = "<anonymous>"
	}
	log.Printf("skipping %s: %s", what, fmt.Sprintln(why...))
}

const primitives = `
// Optional is a type alias for optional values.
//
// # Compatibility Note
//
// In the future, this type will be replaced with an opaque type. This means
// that the type might not be a pointer anymore. Because of this, it is
// recommended to use the methods provided by this package to interact with
// this type.
type Optional[T any] *T

// NewValue returns an optional value from a non-nil value.
func NewValue[T any](v T) Optional[T] { return Optional[T](&v) }

// None returns a nil optional value.
func None[T any]() Optional[T] { return nil }

// Unwrap returns the value of the optional, or panics if the optional is nil.
func (o Optional) Unwrap() T {
	if o == nil {
		panic("attempted to unwrap nil Optional")
	}
	return *o
}

// IsNone returns true if the optional is nil.
func (o Optional) IsNone() bool { return o == nil }
`

// Generate generates the code using the given document.
func (g *Generator) Generate(doc libopenapi.Document) error {
	v3doc, errs := doc.BuildV3Model()
	if errs != nil {
		err := stderrors.Join(errs...)
		return errors.Wrap(err, "failed to build OpenAPI v3 model")
	}

	schemaBytes := parallelMapAttrs(v3doc.Model.Components.Schemas,
		func(name string, proxy *openapibase.SchemaProxy) []byte {
			return g.captured(func(g *Generator) {
				g.error(g.addTopLevelSchema(schemaPath{{Name: name, SchemaProxy: proxy}}))
			})
		})
	schemaBytesIter := orderedMap(schemaBytes)
	schemaBytesIter(func(name string, generated []byte) bool {
		g.output.Write(generated)
		return true
	})

	if g.state.errors.count > 0 {
		err := stderrors.Join(g.state.errors.errors...)
		return errors.Wrapf(err, "encountered %d errors such as", g.state.errors.count)
	}

	return g.output.Flush()
}

func (g *Generator) captured(f func(g *Generator)) []byte {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	generator := *g
	generator.output = w
	f(&generator)

	w.Flush()
	return buf.Bytes()
}

// schemaPath is a path to a specific schema within the document tree.
// The last element is the schema itself.
type schemaPath []schemaLeaf

type schemaLeaf struct {
	*openapibase.SchemaProxy
	Name string
}

// String returns the string representation of the path.
func (p schemaPath) String() string {
	var buf strings.Builder
	for i, schema := range p {
		if i > 0 {
			buf.WriteByte('.')
		}
		buf.WriteString(schema.Name)
	}
	return buf.String()
}

// Parent returns the parent path.
func (p schemaPath) Parent() schemaPath {
	return p[:len(p)-1]
}

// Current returns the current schema.
func (p schemaPath) Current() *openapibase.Schema {
	return p[len(p)-1].SchemaProxy.Schema()
}

// CurrentProxy returns the current schema proxy.
func (p schemaPath) CurrentProxy() *openapibase.SchemaProxy {
	return p[len(p)-1].SchemaProxy
}

// CurrentName returns the name of the current schema.
func (p schemaPath) CurrentName() string {
	return p[len(p)-1].Name
}

// Push pushes a new schema onto the path and returns the new path.
func (p schemaPath) Push(name string, proxy *openapibase.SchemaProxy) schemaPath {
	return append(p, schemaLeaf{Name: name, SchemaProxy: proxy})
}

var integerTypeMap = map[string]string{}

var numberTypeMap = map[string]string{
	"double": "float64",
}

func (g *Generator) addTopLevelSchema(path schemaPath) error {
	b := g.captured(func(g *Generator) {
		fmt.Fprintf(g.output, "type %s ", pascalToGo(path.CurrentName()))
		g.error(g.generateSchema(path))
		fmt.Fprintf(g.output, "\n\n")
	})
	g.state.global.output.Write(b)
	return nil
}

func (g *Generator) generateSchema(path schemaPath) error {
	proxy := path.CurrentProxy()
	if proxy.IsReference() {
		switch ref := proxy.GetReference(); stdpath.Dir(ref) {
		case "#/components/schemas":
			name := stdpath.Base(ref)
			fmt.Fprintf(g.output, "%s", pascalToGo(name))
			return nil
		case "#/components/responses":
			return nil // TODO
		default:
			return fmt.Errorf("unknown reference %q", ref)
		}
	}

	schema := path.Current()

	var primaryType string
	switch len(schema.Type) {
	case 1:
		primaryType = schema.Type[0]
	case 2:
		nullIx := slices.Index(schema.Type, "null")
		if nullIx == -1 {
			return fmt.Errorf("schema %s has more than one type: %q", path, schema.Type)
		}

		fmt.Fprintf(g.output, "*")
		primaryType = schema.Type[1-nullIx]
	default:
		if len(schema.Type) > 0 {
			return fmt.Errorf("schema %s has more than one type: %q", path, schema.Type)
		}
	}

	switch primaryType {
	case "object":
		fmt.Fprintf(g.output, "struct {\n")

		propertiesIter := orderedMap(schema.Properties)
		propertiesIter(func(name string, proxy *openapibase.SchemaProxy) bool {
			fmt.Fprintf(g.output, "\t%s ", snakeToGo(name))

			optional := !slices.Contains(schema.Required, name)
			if optional {
				fmt.Fprintf(g.output, "Optional[")
			}
			g.error(g.generateSchema(path.Push(name, proxy)))
			if optional {
				fmt.Fprintf(g.output, "]")
			}

			jsonKey := name
			if optional {
				jsonKey += ",omitempty"
			}
			fmt.Fprintf(g.output, " `json:%q`", jsonKey)
			fmt.Fprintln(g.output)
			return true
		})

		fmt.Fprintf(g.output, "}")
		return nil
	case "string":
		return g.generateString(path)
	case "integer":
		intType := "int"
		if schema.Format != "" {
			intType = schema.Format
		}
		fmt.Fprintf(g.output, "%s", intType)
		return nil
	case "number":
		fmt.Fprintf(g.output, "float64")
		return nil
	case "boolean":
		fmt.Fprintf(g.output, "bool")
		return nil
	case "null":
		fmt.Fprintf(g.output, "*struct{}")
		return nil
	}

	// TODO: generate Validate() if !opts.anonymous

	switch {
	case schema.AllOf != nil:
		return g.generateAllOf(path, schema.AllOf)
	case schema.AnyOf != nil:
		return g.generateAnyOf(path, schema.AnyOf)
	case schema.OneOf != nil:
		return g.generateOneOf(path, schema.OneOf)
	case schema.Not != nil:
		return fmt.Errorf("unsupported 'not' schema %s", path)
	}

	fmt.Fprintf(g.output, "struct{ /* unsupported %q (%s) */ }", path, schema.Type)
	return nil
}

func (g *Generator) generateString(path schemaPath) error {
	schema := path.Current()
	switch schema.Format {
	case "snowflake":
		log.Println("snowflake type has field name", path.CurrentName())
		log.Println("snowflake type has parent name", path.Parent().CurrentName())

		fmt.Fprintf(g.output, "Snowflake")
	case "date-time":
		fmt.Fprintf(g.output, "time.Time")
	default:
		fmt.Fprintf(g.output, "string")
	}

	return nil
}

func (g *Generator) generateAllOf(path schemaPath, proxies []*openapibase.SchemaProxy) error {
	return nil
}

func (g *Generator) generateAnyOf(path schemaPath, proxies []*openapibase.SchemaProxy) error {
	return nil
}

func (g *Generator) generateOneOf(path schemaPath, proxies []*openapibase.SchemaProxy) error {
	// Special case: Discord gives [{type: null}, T].
	if len(proxies) == 2 {
		resolved, err := resolveSchemas(proxies)
		if err != nil {
			return err
		}

		nullIx := slices.IndexFunc(resolved, func(schema *openapibase.Schema) bool {
			return len(schema.Type) == 1 && schema.Type[0] == "null"
		})
		if nullIx != -1 {
			fmt.Fprintf(g.output, "*")
			g.error(g.generateSchema(path.Push("", proxies[1-nullIx])))
			return nil
		}
	}

	return nil
}

// orderedMap returns a function that iterates over the given map in a
// deterministic order.
//
// Usage:
//
//    for k, v := range orderedMap(m) {
//    	// ...
//    }
//
func orderedMap[K constraints.Ordered, V any](m map[K]V) func(func(K, V) bool) bool {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return func(f func(K, V) bool) bool {
		for _, key := range keys {
			value := m[key]
			if !f(key, value) {
				return false
			}
		}
		return true
	}
}

func parallelMapAttrs[K comparable, V1, V2 any](m map[K]V1, f func(K, V1) V2) map[K]V2 {
	dst := make(map[K]V2, len(m))
	var mu sync.Mutex

	pool := pool.New()
	for k, v := range m {
		k := k
		v := v

		pool.Go(func() {
			v2 := f(k, v)
			mu.Lock()
			dst[k] = v2
			mu.Unlock()
		})
	}

	pool.Wait()
	return dst
}

func mapAttrs[K comparable, V1, V2 any](m map[K]V1, f func(K, V1) V2) map[K]V2 {
	dst := make(map[K]V2, len(m))
	for k, v := range m {
		dst[k] = f(k, v)
	}
	return dst
}

func snakeToGo(s string) string {
	// TODO: vendor strcases.
	return strcases.SnakeToGo(true, s)
}

func pascalToGo(s string) string {
	return strcases.PascalToGo(s)
}

func resolveSchemas(proxies []*openapibase.SchemaProxy) ([]*openapibase.Schema, error) {
	schemas := make([]*openapibase.Schema, 0, len(proxies))
	var errs []error
	for _, proxy := range proxies {
		schema, err := proxy.BuildSchema()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		schemas = append(schemas, schema)
	}
	return schemas, stderrors.Join(errs...)
}

func resolveNamedSchemas(proxies map[string]*openapibase.SchemaProxy) (map[string]*openapibase.Schema, error) {
	schemas := make(map[string]*openapibase.Schema, len(proxies))
	var errs []error
	for name, proxy := range proxies {
		schema, err := proxy.BuildSchema()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		schemas[name] = schema
	}
	return schemas, stderrors.Join(errs...)
}

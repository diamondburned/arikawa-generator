package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	stderrors "errors"
	stdpath "path"

	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-hclog"
	"github.com/pb33f/libopenapi"
	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"libdb.so/arikawa-generator/internal/cmt"
	"libdb.so/arikawa-generator/internal/darkfuckingmagic"
	"libdb.so/arikawa-generator/internal/docread"

	openapibase "github.com/pb33f/libopenapi/datamodel/high/base"
)

type generateState struct {
	sync.Mutex
	generated map[string]string

	errors     []error
	errorCount int

	ctx context.Context
}

func newState(ctx context.Context) *generateState {
	return &generateState{
		generated: map[string]string{},
		ctx:       ctx,
	}
}

func (e *generateState) addError(err error) {
	log := hclog.FromContext(e.ctx)
	log.Error("generate error occured", "error", err)

	e.Lock()
	defer e.Unlock()

	e.errorCount++
	if len(e.errors) < 10 {
		e.errors = append(e.errors, err)
	}
}

func (e *generateState) addGenerated(name, content string) {
	e.Lock()
	o, ok := e.generated[name]
	e.generated[name] = content
	e.Unlock()

	if ok && o != content {
		log := hclog.FromContext(e.ctx)
		log.Error(
			"duplicate name generated in the global scope",
			"name", name,
			"diff", cmp.Diff(content, o))

		e.addError(fmt.Errorf("duplicate name generated in the global scope: %s", name))
	}
}

// Generate generates the code using the given document.
func Generate(doc libopenapi.Document, pkgName string) ([]byte, error) {
	v3doc, errs := doc.BuildV3Model()
	if errs != nil {
		err := stderrors.Join(errs...)
		return nil, errors.Wrap(err, "failed to build OpenAPI v3 model")
	}

	var buf bytes.Buffer
	buf.WriteString("// Code generated by arikawa-generator. DO NOT EDIT.\n\n")
	buf.WriteString("package " + pkgName + "\n\n")

	fmt.Fprintf(&buf, "import %q", optionPkg)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf)

	state := newState(context.TODO())

	// Trim off "Response" if there's no collision.
	for name, schema := range v3doc.Model.Components.Schemas {
		trimmed := strings.TrimSuffix(name, "Response")
		if _, ok := v3doc.Model.Components.Schemas[trimmed]; !ok {
			v3doc.Model.Components.Schemas[trimmed] = schema
			delete(v3doc.Model.Components.Schemas, name)
		}
	}

	parallelMapAttrsInplace(v3doc.Model.Components.Schemas, state.addGenerated,
		func(name string, proxy *openapibase.SchemaProxy) string {
			return generateNamedSchema(state, schemaPath{{Name: name, SchemaProxy: proxy}})
		})
	schemaBytesIter := orderedMap(state.generated)
	schemaBytesIter(func(name, generated string) bool {
		buf.WriteString(generated)
		return true
	})

	if state.errorCount > 0 {
		err := stderrors.Join(state.errors...)
		return nil, errors.Wrapf(err, "encountered %d errors such as", state.errorCount)
	}

	return buf.Bytes(), nil
}

type generator struct {
	output io.Writer
	state  *generateState
}

func generateNamedSchema(state *generateState, path schemaPath) string {
	var b strings.Builder
	g := &generator{output: &b, state: state}
	fmt.Fprintf(g.output, "type %s ", pascalToGo(path.CurrentName()))
	g.generateSchema(path)
	fmt.Fprintf(g.output, "\n\n")
	return b.String()
}

func (g *generator) error(err error) {
	if err != nil {
		g.state.addError(err)
	}
}

func (g *generator) captured(f func(*generator)) string {
	var b strings.Builder
	g2 := &generator{output: &b, state: g.state}
	f(g2)
	return b.String()
}

const (
	pathSchemas   = "#/components/schemas"
	pathResponses = "#/components/responses"
)

var enumNames = NewSet("type", "flags")

func (g *generator) generateSchema(path schemaPath) {
	log := hclog.FromContext(g.state.ctx)
	log.Debug("generating schema", "path", path.String())

	proxy := path.CurrentProxy()
	if proxy.IsReference() && !path.IsRoot() {
		switch ref := proxy.GetReference(); stdpath.Dir(ref) {
		case pathSchemas:
			name := stdpath.Base(ref)
			fmt.Fprintf(g.output, "%s", pascalToGo(name))
			return
		case pathResponses:
			return // TODO
		default:
			g.error(fmt.Errorf("unknown reference %q", ref))
			return
		}
	}

	schema := path.Current()

	ptype, err := extractPrimaryType(schema.Type)
	if err != nil {
		g.error(fmt.Errorf("schema %s has invalid type: %v", path, err))
		return
	}

	if ptype.Nullable {
		fmt.Fprintf(g.output, "*")
	}

	switch ptype.Type {
	case "object":
		g.generateObject(path)
		return
	case "array":
		if !schema.Items.IsA() {
			g.error(fmt.Errorf("schema %s has array type but no items", path))
			return
		}
		fmt.Fprintf(g.output, "[]")
		g.generateSchema(path.Push("_[]", schema.Items.A))
		return
	case "string":
		g.generateString(path)
		return
	case "integer":
		g.generateInteger(path)
		return
	case "number":
		fmt.Fprintf(g.output, "float64")
		return
	case "boolean":
		fmt.Fprintf(g.output, "bool")
		return
	case "null":
		fmt.Fprintf(g.output, "Null")
		return
	}

	switch {
	case schema.AllOf != nil:
		g.generateAllOf(path, schema.AllOf)
		return
	// case schema.AnyOf != nil:
	// 	return g.generateAnyOf(path, schema.AnyOf)
	case schema.OneOf != nil:
		g.generateOneOf(path, schema.OneOf)
		return
	case schema.Not != nil:
		g.error(fmt.Errorf("unsupported 'not' schema %s", path))
		return
	}

	// TODO: generate Validate() if !opts.anonymous
	g.generateUnknown(path)
}

func (g *generator) generateObject(path schemaPath) {
	log := hclog.FromContext(g.state.ctx)
	log.Debug("generating object", "path", path.String())

	schema := path.Current()

	var docCandidateFields map[string]docread.FieldInfo
	// if candidates := calculateTopLikelihood(path, schema); len(candidates) > 0 {
	// 	topCandidate := candidates[0]
	// 	docCandidateFields = docread.ToFieldMap(topCandidate.FieldInfos())
	// }

	fmt.Fprintf(g.output, "struct {\n")

	propertyNames := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		propertyNames = append(propertyNames, name)
	}

	// This OpenAPI library is very funny in the sense that everything we
	// can use actually contains line numbers. So we can sort the properties
	// by line number and get deterministic output.
	propertyLines := make(map[string]int, len(propertyNames))
	for _, name := range propertyNames {
		lowProperty := schema.Properties[name].GoLow()
		propKeyNode := darkfuckingmagic.UnexportedField[*yaml.Node](lowProperty, "kn")
		propertyLines[name] = propKeyNode.Line
	}
	sort.Slice(propertyNames, func(i, j int) bool {
		return propertyLines[propertyNames[i]] < propertyLines[propertyNames[j]]
	})

	for _, name := range propertyNames {
		proxy := schema.Properties[name]
		optional := !slices.Contains(schema.Required, name)

		docField, ok := docCandidateFields[name]
		if ok {
			fmt.Fprint(g.output, cmt.Prettify(snakeToGo(name), docField.Comment, cmt.Opts{
				OriginalName: name,
				Indent:       len(path) - 1,
			}))
		}

		fmt.Fprintf(g.output, "\t%s ", snakeToGo(name))
		if optional {
			fmt.Fprintf(g.output, "option.Optional[")
		}

		t := g.captured(func(g *generator) { g.generateSchema(path.Push(name, proxy)) })
		if optional {
			// Remove pointer from type if the type is already optional.
			t = strings.TrimPrefix(t, "*")
		}

		fmt.Fprint(g.output, t)

		if optional {
			fmt.Fprintf(g.output, "]")
		}

		jsonKey := name
		if optional {
			jsonKey += ",omitempty"
		}
		fmt.Fprintf(g.output, " `json:%q`", jsonKey)
		fmt.Fprintln(g.output)
	}

	fmt.Fprintf(g.output, "}")
}

func (g *generator) generateString(path schemaPath) error {
	log := hclog.FromContext(g.state.ctx)
	log.Debug("generating string", "path", path.String())

	schema := path.Current()
	switch schema.Format {
	case "snowflake":
		fmt.Fprintf(g.output, "%s", g.guessSnowflake(path))
	case "date-time":
		fmt.Fprintf(g.output, "time.Time")
	default:
		fmt.Fprintf(g.output, "string")
		if len(schema.OneOf) > 0 {
			fmt.Fprintln(g.output)
			g.generateConsts(path, schema.OneOf)
		}
	}

	return nil
}

func (g *generator) guessSnowflake(path schemaPath) string {
	// skip
	path = path.PopPrivateLeaves()

	kind, ok := snowflakeFields[path.String()]
	if ok {
		return kind + "ID"
	}

	fieldName := snakeToGo(path.CurrentName())
	parentName := pascalToGo(path.Parent().CurrentName())

	for kind := range snowflakes {
		if false ||
			(kind+"ID" == fieldName) ||
			(kind+"IDs" == fieldName) ||
			(strings.HasSuffix(fieldName, kind)) ||
			(strings.HasSuffix(fieldName, kind+"s")) ||
			(strings.HasSuffix(fieldName, kind+"ID")) ||
			(strings.HasSuffix(fieldName, kind+"IDs")) ||
			(fieldName == "ID" && strings.HasPrefix(parentName, kind)) {

			return kind + "ID"
		}
	}

	log := hclog.FromContext(g.state.ctx)
	log.Debug("unknown snowflake field",
		"path", path.String(),
		"field", fieldName,
		"parent", parentName)

	return "Snowflake"
}

func (g *generator) generateInteger(path schemaPath) {
	log := hclog.FromContext(g.state.ctx)
	log.Debug("generating integer", "path", path.String())

	schema := path.Current()

	var intType string
	if len(schema.AllOf) == 1 && proxyIsGeneratedReference(schema.AllOf[0]) {
		ref := schema.AllOf[0].GetReference()
		intType = pascalToGo(stdpath.Base(ref))
	}
	if intType == "" && enumNames.Has(path.CurrentName()) {
		log := hclog.FromContext(g.state.ctx)
		log.Warn(path.CurrentName()+" is integer but should be enum", "path", path.String())
	}
	if intType == "" && schema.Format != "" {
		intType = schema.Format
	}
	if intType == "" {
		intType = "int"
	}

	fmt.Fprintf(g.output, "%s", intType)

	if len(schema.OneOf) > 0 {
		g.generateConsts(path, schema.OneOf)
	}
}

func (g *generator) generateConsts(path schemaPath, proxies []*openapibase.SchemaProxy) {
	log := hclog.FromContext(g.state.ctx)
	log.Debug("generating consts", "path", path.String())

	fmt.Fprintln(g.output)
	fmt.Fprintln(g.output, "const (")

	prefix := pascalToGo(path.CurrentName())
	prefix = strings.TrimSuffix(prefix, "s") // remove plural

	for _, proxy := range proxies {
		schema := proxy.Schema()

		constName := prefix + constToGo(proxy.Schema().Title)

		constVal, err := schemaConst(schema)
		if err != nil {
			g.error(fmt.Errorf("failed to generate const for %s: %w", path.String(), err))
			continue
		}

		if schema.Description != "" {
			fmt.Fprint(g.output,
				cmt.Prettify(constName, schema.Description, cmt.Opts{Indent: 1}))
		}

		fmt.Fprintf(g.output,
			"\t%s %s = %s\n",
			constName, pascalToGo(path.CurrentName()), constVal)
	}

	fmt.Fprintln(g.output, ")")
}

func (g *generator) generateAllOf(path schemaPath, proxies []*openapibase.SchemaProxy) {
	log := hclog.FromContext(g.state.ctx)
	log.Debug("generating allOf", "path", path.String())

	fmt.Fprintln(g.output, "struct {")
	for i, proxy := range proxies {
		g.generateSchema(path.Push(fmt.Sprintf("_allOf[%d]", i), proxy))
	}
	fmt.Fprintln(g.output, "}")
}

func (g *generator) generateAnyOf(path schemaPath, proxies []*openapibase.SchemaProxy) error {
	log := hclog.FromContext(g.state.ctx)
	log.Debug("generating anyOf", "path", path.String())

	return nil
}

func (g *generator) generateOneOf(path schemaPath, proxies []*openapibase.SchemaProxy) error {
	log := hclog.FromContext(g.state.ctx)

	// Special case: Discord gives [{type: null}, T].
	if len(proxies) == 2 {
		nullIx := slices.IndexFunc(proxies, func(proxy *openapibase.SchemaProxy) bool {
			schema := proxy.Schema()
			return slices.Equal(schema.Type, []string{"null"})
		})
		if nullIx != -1 {
			fmt.Fprintf(g.output, "*")
			i := 1 - nullIx
			g.generateSchema(path.Push(fmt.Sprintf("_oneOf[%d]", i), proxies[i]))
			return nil
		}
	}

	var name string
	if path.CurrentIsExported() {
		name = pascalToGo(path.CurrentName())
	} else {
		for _, part := range path {
			if part.IsPrivate() {
				continue
			}
			name += strcases.Go(part.Name)
		}
	}

	log.Debug("generating oneOf", "path", path, "name", name)

	// Generate this as a reference to a type, but we'll generate the type
	// globally.
	fmt.Fprint(g.output, name)

	content := g.captured(func(g *generator) { g.generateNamedOneOf(path, name, proxies) })
	if path.IsRoot() {
		fmt.Fprint(g.output, content)
	} else {
		g.state.addGenerated(name, content)
	}

	return nil
}

func (g *generator) generateNamedOneOf(path schemaPath, unionName string, proxies []*openapibase.SchemaProxy) {
	names := make([]string, len(proxies))

	var comment strings.Builder
	comment.WriteString("is a union of the following types:\n")
	for i, proxy := range proxies {
		if proxyIsGeneratedReference(proxy) {
			names[i] = pascalToGo(stdpath.Base(proxy.GetReference()))
			goto named
		}

		{
			schema := proxy.Schema()

			ptype, err := extractPrimaryType(schema.Type)
			if err == nil {
				names[i] = fmt.Sprintf("%s%s", unionName, snakeToGo(ptype.Type))
				goto named
			}
		}

		names[i] = fmt.Sprintf("%s%d", unionName, i)
	named:
		comment.WriteString(fmt.Sprintf("  - [%s]\n", names[i]))
	}
	comment.WriteString("\n")

	fmt.Fprintf(g.output, cmt.Prettify(unionName, comment.String(), cmt.Opts{NoWrap: true}))
	fmt.Fprintf(g.output, "type %s interface {\n", unionName)
	fmt.Fprintf(g.output, "  is%s()\n", unionName)
	fmt.Fprintf(g.output, "}\n\n")

	for _, name := range names {
		fmt.Fprintf(g.output, "func (%s) is%s() {}\n", name, unionName)
	}
	fmt.Fprintln(g.output)

	// Generate all inlined types.
	for i, proxy := range proxies {
		if proxyIsGeneratedReference(proxy) {
			continue
		}

		fmt.Fprintf(g.output, "type %s ", names[i])
		g.generateSchema(path.Push(names[i], proxy))
		fmt.Fprintln(g.output)
		fmt.Fprintln(g.output)
	}
}

type primaryType struct {
	Type     string
	Nullable bool
}

func extractPrimaryType(types []string) (primaryType, error) {
	switch len(types) {
	case 0:
		return primaryType{}, nil
	case 1:
		return primaryType{Type: types[0]}, nil
	case 2:
		nullIx := slices.Index(types, "null")
		if nullIx == -1 {
			return primaryType{}, errors.New("schema has more than one type")
		}
		return primaryType{Type: types[1-nullIx], Nullable: true}, nil
	default:
		return primaryType{}, errors.New("schema has more than one type")
	}
}

func (g *generator) generateUnknown(path schemaPath) {
	log := hclog.FromContext(g.state.ctx)
	log.Warn("unknown schema type", "path", path.String())

	schema := path.Current()
	fmt.Fprintf(g.output, "struct{ /* %v */ }", schema.Type)
}

func proxyIsGeneratedReference(proxy *openapibase.SchemaProxy) bool {
	if !proxy.IsReference() {
		return false
	}
	ref := proxy.GetReference()
	return stdpath.Dir(ref) == pathSchemas
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
	add := func(k K, v V2) {
		mu.Lock()
		dst[k] = v
		mu.Unlock()
	}
	parallelMapAttrsInplace(m, add, f)
	return dst
}

func parallelMapAttrsInplace[K comparable, V1, V2 any](m map[K]V1, add func(K, V2), f func(K, V1) V2) {
	pool := pool.New().WithMaxGoroutines(numWorkers)
	for k, v := range m {
		k := k
		v := v
		pool.Go(func() {
			v2 := f(k, v)
			add(k, v2)
		})
	}
	pool.Wait()
}

func mapAttrs[K comparable, V1, V2 any](m map[K]V1, f func(K, V1) V2) map[K]V2 {
	dst := make(map[K]V2, len(m))
	for k, v := range m {
		dst[k] = f(k, v)
	}
	return dst
}

var illegalDestroyer = strings.NewReplacer(
	"[", "",
	"]", "",
	"-", "",
	" ", "",
)

func snakeToGo(s string) string {
	// TODO: vendor strcases.
	s = illegalDestroyer.Replace(s)
	return strcases.SnakeToGo(true, s)
}

func constToGo(s string) string {
	s = strings.ToLower(s)
	return snakeToGo(s)
}

func pascalToGo(s string) string {
	s = illegalDestroyer.Replace(s)
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

// schemaConst returns the "const" field of the given schema, if any. The
// OpenAPI library doesn't support this field, so we have to do it ourselves.
func schemaConst(schema *openapibase.Schema) (string, error) {
	var v struct {
		Const any `json:"const" yaml:"const"`
	}

	node := schema.GoLow().ParentProxy.GetValueNode()
	if err := node.Decode(&v); err != nil {
		return "", err
	}

	// TODO: be less lazy
	b, err := json.Marshal(v.Const)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func isExported(s string) bool {
	if s == "" {
		return false
	}
	l := strcases.FirstLetter(s)
	return l == strings.ToUpper(l)
}

package main

import (
	"strings"

	openapibase "github.com/pb33f/libopenapi/datamodel/high/base"
)

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

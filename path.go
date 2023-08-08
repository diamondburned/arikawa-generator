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

func (l schemaLeaf) IsPrivate() bool {
	return l.Name == "" || strings.HasPrefix(l.Name, "_")
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
	leaf := p[len(p)-1]
	if leaf.IsPrivate() {
		return ""
	}
	return leaf.Name
}

// CurrentIsPrivate returns whether the current schema is private, meaning it is
// generated as a pseudo field and not a real one.
func (p schemaPath) CurrentIsPrivate() bool {
	leaf := p[len(p)-1]
	return leaf.IsPrivate()
}

// CurrentIsExported returns whether the current schema is exported.
func (p schemaPath) CurrentIsExported() bool {
	return isExported(p.PopPrivateLeaves().CurrentName())
}

// PopPrivateLeaves pops all private leaves from the path and returns the new
// path.
func (p schemaPath) PopPrivateLeaves() schemaPath {
	for len(p) > 0 && p[len(p)-1].IsPrivate() {
		p = p[:len(p)-1]
	}
	return p
}

// Push pushes a new schema onto the path and returns the new path.
func (p schemaPath) Push(name string, proxy *openapibase.SchemaProxy) schemaPath {
	return append(p, schemaLeaf{Name: name, SchemaProxy: proxy})
}

// IsRoot returns whether the path is the root path.
func (p schemaPath) IsRoot() bool { return len(p) <= 1 }

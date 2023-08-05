package denodoc

import (
	"encoding/json"
	"fmt"
)

// export type DocNode =
//   | DocNodeModuleDoc
//   | DocNodeFunction
//   | DocNodeVariable
//   | DocNodeEnum
//   | DocNodeClass
//   | DocNodeTypeAlias
//   | DocNodeNamespace
//   | DocNodeInterface
//   | DocNodeImport;

type DocNodeKind string

const (
	DocNodeKindModuleDoc DocNodeKind = "moduleDoc"
	DocNodeKindFunction  DocNodeKind = "function"
	DocNodeKindVariable  DocNodeKind = "variable"
	DocNodeKindEnum      DocNodeKind = "enum"
	DocNodeKindClass     DocNodeKind = "class"
	DocNodeKindTypeAlias DocNodeKind = "typeAlias"
	DocNodeKindNamespace DocNodeKind = "namespace"
	DocNodeKindInterface DocNodeKind = "interface"
	DocNodeKindImport    DocNodeKind = "import"
)

type DocNode struct{ Value IDocNode }
type IDocNode interface{ Kind() DocNodeKind }

func (DocNodeModuleDoc) Kind() DocNodeKind { return DocNodeKindModuleDoc }
func (DocNodeFunction) Kind() DocNodeKind  { return DocNodeKindFunction }
func (DocNodeVariable) Kind() DocNodeKind  { return DocNodeKindVariable }
func (DocNodeEnum) Kind() DocNodeKind      { return DocNodeKindEnum }
func (DocNodeClass) Kind() DocNodeKind     { return DocNodeKindClass }
func (DocNodeTypeAlias) Kind() DocNodeKind { return DocNodeKindTypeAlias }
func (DocNodeNamespace) Kind() DocNodeKind { return DocNodeKindNamespace }
func (DocNodeInterface) Kind() DocNodeKind { return DocNodeKindInterface }
func (DocNodeImport) Kind() DocNodeKind    { return DocNodeKindImport }

func (d *DocNode) UnmarshalJSON(b []byte) error {
	var kind struct{ Kind DocNodeKind }
	if err := json.Unmarshal(b, &kind); err != nil {
		return fmt.Errorf("unmarshal DocNodeKind: %w", err)
	}

	var err error

	switch kind.Kind {
	case DocNodeKindModuleDoc:
		var v DocNodeModuleDoc
		err = json.Unmarshal(b, &v)
		d.Value = v
	case DocNodeKindFunction:
		var v DocNodeFunction
		err = json.Unmarshal(b, &v)
		d.Value = v
	case DocNodeKindVariable:
		var v DocNodeVariable
		err = json.Unmarshal(b, &v)
		d.Value = v
	case DocNodeKindEnum:
		var v DocNodeEnum
		err = json.Unmarshal(b, &v)
		d.Value = v
	case DocNodeKindClass:
		var v DocNodeClass
		err = json.Unmarshal(b, &v)
		d.Value = v
	case DocNodeKindTypeAlias:
		var v DocNodeTypeAlias
		err = json.Unmarshal(b, &v)
		d.Value = v
	case DocNodeKindNamespace:
		var v DocNodeNamespace
		err = json.Unmarshal(b, &v)
		d.Value = v
	case DocNodeKindInterface:
		var v DocNodeInterface
		err = json.Unmarshal(b, &v)
		d.Value = v
	case DocNodeKindImport:
		var v DocNodeImport
		err = json.Unmarshal(b, &v)
		d.Value = v
	default:
		return fmt.Errorf("unknown DocNodeKind: %s", kind.Kind)
	}

	return err
}

func (d DocNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind DocNodeKind
		IDocNode
	}{
		Kind:     d.Value.Kind(),
		IDocNode: d.Value,
	})
}

//
// /** Indicates how the documentation node was declared. `"private"` indicates
//  * the node is un-exported. `"export"` indicates it is exported from the current
//  * module. `"declare"` indicates that it is a type only declaration. */
// export type DeclarationKind = "private" | "export" | "declare";

type DeclarationKind string

const (
	DeclarationKindPrivate DeclarationKind = "private"
	DeclarationKindExport  DeclarationKind = "export"
	DeclarationKindDeclare DeclarationKind = "declare"
)

//
// interface DocNodeBase {
//   kind: DocNodeKind;
//   name: string;
//   location: Location;
//   declarationKind: DeclarationKind;
//   jsDoc?: JsDoc;
// }

type DocNodeBase struct {
	Name            string
	DeclarationKind DeclarationKind
}

//
// export interface DocNodeModuleDoc extends DocNodeBase {
//   kind: "moduleDoc";
//   jsDoc: JsDoc;
// }

type DocNodeModuleDoc struct {
	DocNodeBase
}

//
// export interface DocNodeFunction extends DocNodeBase {
//   kind: "function";
//   functionDef: FunctionDef;
// }

type DocNodeFunction struct {
	DocNodeBase
	FunctionDef FunctionDef
}

//
// export interface DocNodeVariable extends DocNodeBase {
//   kind: "variable";
//   variableDef: VariableDef;
// }

type DocNodeVariable struct {
	DocNodeBase
	VariableDef VariableDef
}

//
// export interface DocNodeEnum extends DocNodeBase {
//   kind: "enum";
//   enumDef: EnumDef;
// }

type DocNodeEnum struct {
	DocNodeBase
	EnumDef EnumDef
}

//
// export interface DocNodeClass extends DocNodeBase {
//   kind: "class";
//   classDef: ClassDef;
// }

type DocNodeClass struct {
	DocNodeBase
	ClassDef ClassDef
}

//
// export interface DocNodeTypeAlias extends DocNodeBase {
//   kind: "typeAlias";
//   typeAliasDef: TypeAliasDef;
// }

type DocNodeTypeAlias struct {
	DocNodeBase
	TypeAliasDef TypeAliasDef
}

//
// export interface DocNodeNamespace extends DocNodeBase {
//   kind: "namespace";
//   namespaceDef: NamespaceDef;
// }

type DocNodeNamespace struct {
	DocNodeBase
	NamespaceDef NamespaceDef
}

//
// export interface DocNodeInterface extends DocNodeBase {
//   kind: "interface";
//   interfaceDef: InterfaceDef;
// }

type DocNodeInterface struct {
	DocNodeBase
	InterfaceDef InterfaceDef
}

//
// export interface DocNodeImport extends DocNodeBase {
//   kind: "import";
//   importDef: ImportDef;
// }

type DocNodeImport struct {
	DocNodeBase
	ImportDef ImportDef
}

//
// export type Accessibility = "public" | "protected" | "private";
//

type Accessibility string

const (
	AccessibilityPublic    Accessibility = "public"
	AccessibilityProtected Accessibility = "protected"
	AccessibilityPrivate   Accessibility = "private"
)

// export interface ClassDef {
//   isAbstract: boolean;
//   constructors: ClassConstructorDef[];
//   properties: ClassPropertyDef[];
//   indexSignatures: ClassIndexSignatureDef[];
//   methods: ClassMethodDef[];
//   extends?: string;
//   implements: TsTypeDef[];
//   typeParams: TsTypeParamDef[];
//   superTypeParams: TsTypeDef[];
//   decorators?: DecoratorDef[];
// }

type ClassDef struct {
	IsAbstract      bool
	Constructors    []ClassConstructorDef
	Properties      []ClassPropertyDef
	IndexSignatures []ClassIndexSignatureDef
	Methods         []ClassMethodDef
	Extends         string
	Implements      []TsTypeDef
	TypeParams      []TsTypeParamDef
	SuperTypeParams []TsTypeDef
	Decorators      []DecoratorDef
}

//
// export type ClassConstructorParamDef = ParamDef & {
//   accessibility?: Accessibility;
//   isOverride?: boolean;
//   readonly?: boolean;
// };

type ClassConstructorParamDef struct {
	ParamDef
	Accessibility Accessibility
	IsOverride    bool
	ReadOnly      bool
}

//
// export interface ClassConstructorDef {
//   jsDoc?: JsDoc;
//   accessibility?: Accessibility;
//   isOptional?: boolean;
//   hasBody?: boolean;
//   name: string;
//   params: ClassConstructorParamDef[];
//   location: Location;
// }

type ClassConstructorDef struct {
	Accessibility Accessibility
	IsOptional    bool
	HasBody       bool
	Name          string
	Params        []ClassConstructorParamDef
}

//
// export interface ClassIndexSignatureDef {
//   readonly: boolean;
//   params: ParamDef[];
//   tsType?: TsTypeDef;
// }
//

type ClassIndexSignatureDef struct {
	ReadOnly bool
	Params   []ParamDef
	TsType   TsTypeDef
}

// export interface ClassMethodDef {
//   jsDoc?: JsDoc;
//   accessibility?: Accessibility;
//   optional: boolean;
//   isAbstract: boolean;
//   isStatic: boolean;
//   isOverride?: boolean;
//   name: string;
//   kind: MethodKind;
//   functionDef: FunctionDef;
//   location: Location;
// }

type ClassMethodDef struct {
	Accessibility Accessibility
	Optional      bool
	IsAbstract    bool
	IsStatic      bool
	IsOverride    bool
	Name          string
	Kind          MethodKind
	FunctionDef   FunctionDef
}

//
// export interface ClassPropertyDef {
//   jsDoc?: JsDoc;
//   tsType?: TsTypeDef;
//   readonly: boolean;
//   accessibility?: Accessibility;
//   optional: boolean;
//   isAbstract: boolean;
//   isStatic: boolean;
//   isOverride?: boolean;
//   name: string;
//   decorators?: DecoratorDef[];
//   location: Location;
// }

type ClassPropertyDef struct {
	TsType        TsTypeDef
	ReadOnly      bool
	Accessibility Accessibility
	Optional      bool
	IsAbstract    bool
	IsStatic      bool
	IsOverride    bool
	Name          string
	Decorators    []DecoratorDef
}

//
// export interface DecoratorDef {
//   name: string;
//   args?: string[];
//   location: Location;
// }

type DecoratorDef struct {
	Name string
	Args []string
}

//
// export interface EnumDef {
//   members: EnumMemberDef[];
// }

type EnumDef struct {
	Members []EnumMemberDef
}

//
// export interface EnumMemberDef {
//   name: string;
//   init?: TsTypeDef;
//   jsDoc?: JsDoc;
//   location: Location;
// }

type EnumMemberDef struct {
	Name string
	Init TsTypeDef
}

//
// export interface FunctionDef {
//   params: ParamDef[];
//   returnType?: TsTypeDef;
//   hasBody?: boolean;
//   isAsync: boolean;
//   isGenerator: boolean;
//   typeParams: TsTypeParamDef[];
//   decorators?: DecoratorDef[];
// }

type FunctionDef struct {
	Params      []ParamDef
	ReturnType  TsTypeDef
	HasBody     bool
	IsAsync     bool
	IsGenerator bool
	TypeParams  []TsTypeParamDef
	Decorators  []DecoratorDef
}

//
// export interface ImportDef {
//   src: string;
//   imported?: string;
// }

type ImportDef struct {
	Src      string
	Imported string
}

//
// export interface InterfaceDef {
//   extends: TsTypeDef[];
//   methods: InterfaceMethodDef[];
//   properties: InterfacePropertyDef[];
//   callSignatures: InterfaceCallSignatureDef[];
//   indexSignatures: InterfaceIndexSignatureDef[];
//   typeParams: TsTypeParamDef[];
// }

type InterfaceDef struct {
	Extends         []TsTypeDef
	Methods         []InterfaceMethodDef
	Properties      []InterfacePropertyDef
	CallSignatures  []InterfaceCallSignatureDef
	IndexSignatures []InterfaceIndexSignatureDef
	TypeParams      []TsTypeParamDef
}

//
// export interface InterfaceCallSignatureDef {
//   location: Location;
//   jsDoc?: JsDoc;
//   params: ParamDef[];
//   tsType?: TsTypeDef;
//   typeParams: TsTypeParamDef[];
// }

type InterfaceCallSignatureDef struct {
	Params    []ParamDef
	TsType    TsTypeDef
	TypeParam []TsTypeParamDef
}

//
// export interface InterfaceIndexSignatureDef {
//   readonly: boolean;
//   params: ParamDef[];
//   tsType?: TsTypeDef;
// }

type InterfaceIndexSignatureDef struct {
	ReadOnly bool
	Params   []ParamDef
	TsType   TsTypeDef
}

//
// export interface InterfaceMethodDef {
//   name: string;
//   kind: MethodKind;
//   location: Location;
//   jsDoc?: JsDoc;
//   computed?: boolean;
//   optional: boolean;
//   params: ParamDef[];
//   returnType?: TsTypeDef;
//   typeParams: TsTypeParamDef[];
// }

type InterfaceMethodDef struct {
	Name       string
	Kind       MethodKind
	Computed   bool
	Optional   bool
	Params     []ParamDef
	ReturnType TsTypeDef
	TypeParams []TsTypeParamDef
}

//
// export interface InterfacePropertyDef {
//   name: string;
//   location: Location;
//   jsDoc?: JsDoc;
//   params: ParamDef[];
//   readonly?: boolean;
//   computed: boolean;
//   optional: boolean;
//   tsType?: TsTypeDef;
//   typeParams: TsTypeParamDef[];
// }

type InterfacePropertyDef struct {
	Name       string
	Params     []ParamDef
	ReadOnly   bool
	Computed   bool
	Optional   bool
	TsType     TsTypeDef
	TypeParams []TsTypeParamDef
}

//
// export interface LiteralCallSignatureDef {
//   params: ParamDef[];
//   tsType?: TsTypeDef;
//   typeParams: TsTypeParamDef[];
// }

type LiteralCallSignatureDef struct {
	Params    []ParamDef
	TsType    TsTypeDef
	TypeParam []TsTypeParamDef
}

//
// export type LiteralDefKind =
//   | "number"
//   | "string"
//   | "template"
//   | "boolean"
//   | "bigInt";

type LiteralDefKind string

const (
	LiteralDefKindNumber   LiteralDefKind = "number"
	LiteralDefKindString   LiteralDefKind = "string"
	LiteralDefKindTemplate LiteralDefKind = "template"
	LiteralDefKindBoolean  LiteralDefKind = "boolean"
	LiteralDefKindBigInt   LiteralDefKind = "bigInt"
)

//
// export type LiteralDef =
//   | LiteralDefNumber
//   | LiteralDefBigInt
//   | LiteralDefString
//   | LiteralDefTemplate
//   | LiteralDefBoolean;

type LiteralDef struct{ Value ILiteralDef }
type ILiteralDef interface{ Kind() LiteralDefKind }

func (LiteralDefNumber) Kind() LiteralDefKind   { return LiteralDefKindNumber }
func (LiteralDefBigInt) Kind() LiteralDefKind   { return LiteralDefKindBigInt }
func (LiteralDefString) Kind() LiteralDefKind   { return LiteralDefKindString }
func (LiteralDefTemplate) Kind() LiteralDefKind { return LiteralDefKindTemplate }
func (LiteralDefBoolean) Kind() LiteralDefKind  { return LiteralDefKindBoolean }

func (l *LiteralDef) UnmarshalJSON(data []byte) error {
	var v struct {
		Kind LiteralDefKind `json:"kind"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	var err error

	switch v.Kind {
	case LiteralDefKindNumber:
		var vv LiteralDefNumber
		err = json.Unmarshal(data, &vv)
		l.Value = vv
	case LiteralDefKindBigInt:
		var vv LiteralDefBigInt
		err = json.Unmarshal(data, &vv)
		l.Value = vv
	case LiteralDefKindString:
		var vv LiteralDefString
		err = json.Unmarshal(data, &vv)
		l.Value = vv
	case LiteralDefKindTemplate:
		var vv LiteralDefTemplate
		err = json.Unmarshal(data, &vv)
		l.Value = vv
	case LiteralDefKindBoolean:
		var vv LiteralDefBoolean
		err = json.Unmarshal(data, &vv)
		l.Value = vv
	default:
		return fmt.Errorf("unknown LiteralDefKind %q", v.Kind)
	}

	return err
}

//
// interface LiteralDefBase {
//   kind: LiteralDefKind;
// }
//
// export interface LiteralDefNumber extends LiteralDefBase {
//   kind: "number";
//   number: number;
// }

type LiteralDefNumber struct {
	Number float64
}

//
// export interface LiteralDefBigInt extends LiteralDefBase {
//   kind: "bigInt";
//   string: string;
// }

type LiteralDefBigInt struct {
	String string
}

//
// export interface LiteralDefString extends LiteralDefBase {
//   kind: "string";
//   string: string;
// }

type LiteralDefString struct {
	String string
}

//
// export interface LiteralDefTemplate extends LiteralDefBase {
//   kind: "template";
//   tsTypes: TsTypeDef[];
// }

type LiteralDefTemplate struct {
	TsTypes []TsTypeDef
}

//
// export interface LiteralDefBoolean extends LiteralDefBase {
//   kind: "boolean";
//   boolean: boolean;
// }

type LiteralDefBoolean struct {
	Boolean bool
}

//
// export interface LiteralIndexSignatureDef {
//   readonly: boolean;
//   params: ParamDef[];
//   tsType?: TsTypeDef;
// }

type LiteralIndexSignatureDef struct {
	ReadOnly bool
	Params   []ParamDef
	TsType   TsTypeDef
}

//
// export interface LiteralMethodDef {
//   name: string;
//   kind: MethodKind;
//   params: ParamDef[];
//   computed?: boolean;
//   optional: boolean;
//   returnType?: TsTypeDef;
//   typeParams: TsTypeParamDef[];
// }

type LiteralMethodDef struct {
	Name       string
	Kind       MethodKind
	Computed   bool
	Optional   bool
	Params     []ParamDef
	ReturnType TsTypeDef
	TypeParams []TsTypeParamDef
}

//
// export interface LiteralPropertyDef {
//   name: string;
//   params: ParamDef[];
//   readonly?: boolean;
//   computed: boolean;
//   optional: boolean;
//   tsType?: TsTypeDef;
//   typeParams: TsTypeParamDef[];
// }

type LiteralPropertyDef struct {
	Name       string
	Params     []ParamDef
	ReadOnly   bool
	Computed   bool
	Optional   bool
	TsType     TsTypeDef
	TypeParams []TsTypeParamDef
}

//
// export interface Location {
//   filename: string;
//   line: number;
//   col: number;
// }
//
// export type MethodKind = "method" | "getter" | "setter";

type MethodKind string

const (
	MethodKindMethod MethodKind = "method"
	MethodKindGetter MethodKind = "getter"
	MethodKindSetter MethodKind = "setter"
)

//
// export interface NamespaceDef {
//   elements: DocNode[];
// }

type NamespaceDef struct {
	Elements []DocNode
}

//
// export type ObjectPatPropDef =
//   | ObjectPatPropAssignDef
//   | ObjectPatPropKeyValueDef
//   | ObjectPatPropRestDef;

type ObjectPatPropDefKind string

const (
	ObjectPatPropDefKindAssign   ObjectPatPropDefKind = "assign"
	ObjectPatPropDefKindKeyValue ObjectPatPropDefKind = "keyValue"
	ObjectPatPropDefKindRest     ObjectPatPropDefKind = "rest"
)

type ObjectPatPropDef struct{ Value IObjectPatPropDef }
type IObjectPatPropDef interface{ Kind() ObjectPatPropDefKind }

func (ObjectPatPropAssignDef) Kind() ObjectPatPropDefKind   { return ObjectPatPropDefKindAssign }
func (ObjectPatPropKeyValueDef) Kind() ObjectPatPropDefKind { return ObjectPatPropDefKindKeyValue }
func (ObjectPatPropRestDef) Kind() ObjectPatPropDefKind     { return ObjectPatPropDefKindRest }

func (o *ObjectPatPropDef) UnmarshalJSON(b []byte) error {
	var v struct {
		Kind ObjectPatPropDefKind `json:"kind"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	var err error

	switch v.Kind {
	case ObjectPatPropDefKindAssign:
		var vv ObjectPatPropAssignDef
		err = json.Unmarshal(b, &vv)
		o.Value = vv
	case ObjectPatPropDefKindKeyValue:
		var vv ObjectPatPropKeyValueDef
		err = json.Unmarshal(b, &vv)
		o.Value = vv
	case ObjectPatPropDefKindRest:
		var vv ObjectPatPropRestDef
		err = json.Unmarshal(b, &vv)
		o.Value = vv
	default:
		return fmt.Errorf("unknown ObjectPatPropDefKind %q", v.Kind)
	}

	return err
}

//
// export interface ObjectPatPropAssignDef {
//   kind: "assign";
//   key: string;
//   value?: string;
// }

type ObjectPatPropAssignDef struct {
	Key   string
	Value string
}

//
// export interface ObjectPatPropKeyValueDef {
//   kind: "keyValue";
//   key: string;
//   value: ParamDef;
// }

type ObjectPatPropKeyValueDef struct {
	Key   string
	Value ParamDef
}

// export interface ObjectPatPropRestDef {
//   kind: "rest";
//   arg: ParamDef;
// }

type ObjectPatPropRestDef struct {
	Arg ParamDef
}

//
// export type ParamDef =
//   | ParamArrayDef
//   | ParamAssignDef
//   | ParamIdentifierDef
//   | ParamObjectDef
//   | ParamRestDef;

type ParamDefKind string

const (
	ParamDefKindArray      ParamDefKind = "array"
	ParamDefKindAssign     ParamDefKind = "assign"
	ParamDefKindIdentifier ParamDefKind = "identifier"
	ParamDefKindObject     ParamDefKind = "object"
	ParamDefKindRest       ParamDefKind = "rest"
)

type ParamDef struct{ Value IParamDef }
type IParamDef interface{ Kind() ParamDefKind }

func (ParamArrayDef) Kind() ParamDefKind      { return ParamDefKindArray }
func (ParamAssignDef) Kind() ParamDefKind     { return ParamDefKindAssign }
func (ParamIdentifierDef) Kind() ParamDefKind { return ParamDefKindIdentifier }
func (ParamObjectDef) Kind() ParamDefKind     { return ParamDefKindObject }
func (ParamRestDef) Kind() ParamDefKind       { return ParamDefKindRest }

func (p *ParamDef) UnmarshalJSON(b []byte) error {
	var v struct {
		Kind ParamDefKind `json:"kind"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	var err error

	switch v.Kind {
	case ParamDefKindArray:
		var vv ParamArrayDef
		err = json.Unmarshal(b, &vv)
		p.Value = vv
	case ParamDefKindAssign:
		var vv ParamAssignDef
		err = json.Unmarshal(b, &vv)
		p.Value = vv
	case ParamDefKindIdentifier:
		var vv ParamIdentifierDef
		err = json.Unmarshal(b, &vv)
		p.Value = vv
	case ParamDefKindObject:
		var vv ParamObjectDef
		err = json.Unmarshal(b, &vv)
		p.Value = vv
	case ParamDefKindRest:
		var vv ParamRestDef
		err = json.Unmarshal(b, &vv)
		p.Value = vv
	default:
		return fmt.Errorf("unknown ParamDefKind %q", v.Kind)
	}

	return err
}

//
// export interface ParamArrayDef {
//   kind: "array";
//   elements: (ParamDef | undefined)[];
//   optional: boolean;
//   decorators?: DecoratorDef[];
//   tsType?: TsTypeDef;
// }

type ParamArrayDef struct {
	Elements   []ParamDef
	Optional   bool
	Decorators []DecoratorDef
	TsType     TsTypeDef
}

//
// export interface ParamAssignDef {
//   kind: "assign";
//   left: ParamDef;
//   right: string;
//   decorators?: DecoratorDef[];
//   tsType?: TsTypeDef;
// }

type ParamAssignDef struct {
	Left       ParamDef
	Right      string
	Decorators []DecoratorDef
	TsType     TsTypeDef
}

//
// export interface ParamIdentifierDef {
//   kind: "identifier";
//   name: string;
//   optional: boolean;
//   decorators?: DecoratorDef[];
//   tsType?: TsTypeDef;
// }

type ParamIdentifierDef struct {
	Name       string
	Optional   bool
	Decorators []DecoratorDef
	TsType     TsTypeDef
}

//
// export interface ParamObjectDef {
//   kind: "object";
//   props: ObjectPatPropDef[];
//   optional: boolean;
//   decorators?: DecoratorDef[];
//   tsType?: TsTypeDef;
// }
//
// export interface ParamRestDef {
//   kind: "rest";
//   arg: ParamDef;
//   decorators?: DecoratorDef[];
//   tsType?: TsTypeDef;
// }
//
// export type TruePlusMinus = true | "+" | "-";
//
// export interface TsConditionalDef {
//   checkType: TsTypeDef;
//   extendsType: TsTypeDef;
//   trueType: TsTypeDef;
//   falseType: TsTypeDef;
// }
//
// export interface TsFnOrConstructorDef {
//   constructor: boolean;
//   tsType: TsTypeDef;
//   params: ParamDef[];
//   typeParams: TsTypeParamDef[];
// }
//
// export interface TsImportTypeDef {
//   specifier: string;
//   qualifier?: string;
//   typeParams?: TsTypeDef[];
// }
//
// export interface TsIndexedAccessDef {
//   readonly: boolean;
//   objType: TsTypeDef;
//   indexType: TsTypeDef;
// }
//
// export interface TsInferDef {
//   typeParam: TsTypeParamDef;
// }
//
// export interface TsMappedTypeDef {
//   readonly?: TruePlusMinus;
//   typeParam: TsTypeParamDef;
//   nameType?: TsTypeDef;
//   optional?: TruePlusMinus;
//   tsType?: TsTypeDef;
// }
//
// export interface TsTypeLiteralDef {
//   methods: LiteralMethodDef[];
//   properties: LiteralPropertyDef[];
//   callSignatures: LiteralCallSignatureDef[];
//   indexSignatures: LiteralIndexSignatureDef[];
// }
//
// export interface TsTypeOperatorDef {
//   operator: string;
//   tsType: TsTypeDef;
// }
//
// export interface TsTypeParamDef {
//   name: string;
//   constraint?: TsTypeDef;
//   default?: TsTypeDef;
// }
//
// export interface TsTypePredicateDef {
//   asserts: boolean;
//   param: { type: "this" | "identifier"; name?: string };
//   type?: TsTypeDef;
// }
//
// export type TsTypeDef =
//   | TsTypeKeywordDef
//   | TsTypeDefLiteral
//   | TsTypeTypeRefDef
//   | TsTypeUnionDef
//   | TsTypeIntersectionDef
//   | TsTypeArrayDef
//   | TsTypeTupleDef
//   | TsTypeTypeOperatorDef
//   | TsTypeParenthesizedDef
//   | TsTypeRestDef
//   | TsTypeOptionalDef
//   | TsTypeQueryDef
//   | TsTypeThisDef
//   | TsTypeFnOrConstructorDef
//   | TsTypeConditionalDef
//   | TsTypeImportTypeDef
//   | TsTypeInferDef
//   | TsTypeIndexedAccessDef
//   | TsTypeMappedDef
//   | TsTypeTypeLiteralDef
//   | TsTypeTypePredicateDef;
//
// interface TsTypeDefBase {
//   repr: string;
//   kind: TsTypeDefKind;
// }
//
// export interface TsTypeKeywordDef extends TsTypeDefBase {
//   kind: "keyword";
//   keyword: string;
// }
//
// export interface TsTypeDefLiteral extends TsTypeDefBase {
//   kind: "literal";
//   literal: LiteralDef;
// }
//
// export interface TsTypeTypeRefDef extends TsTypeDefBase {
//   kind: "typeRef";
//   typeRef: TsTypeRefDef;
// }
//
// export interface TsTypeUnionDef extends TsTypeDefBase {
//   kind: "union";
//   union: TsTypeDef[];
// }
//
// export interface TsTypeIntersectionDef extends TsTypeDefBase {
//   kind: "intersection";
//   intersection: TsTypeDef[];
// }
//
// export interface TsTypeArrayDef extends TsTypeDefBase {
//   kind: "array";
//   array: TsTypeDef;
// }
//
// export interface TsTypeTupleDef extends TsTypeDefBase {
//   kind: "tuple";
//   tuple: TsTypeDef[];
// }
//
// export interface TsTypeTypeOperatorDef extends TsTypeDefBase {
//   kind: "typeOperator";
//   typeOperator: TsTypeOperatorDef;
// }
//
// export interface TsTypeParenthesizedDef extends TsTypeDefBase {
//   kind: "parenthesized";
//   parenthesized: TsTypeDef;
// }
//
// export interface TsTypeRestDef extends TsTypeDefBase {
//   kind: "rest";
//   rest: TsTypeDef;
// }
//
// export interface TsTypeOptionalDef extends TsTypeDefBase {
//   kind: "optional";
//   optional: TsTypeDef;
// }
//
// export interface TsTypeQueryDef extends TsTypeDefBase {
//   kind: "typeQuery";
//   typeQuery: string;
// }
//
// export interface TsTypeThisDef extends TsTypeDefBase {
//   kind: "this";
//   this: boolean;
// }
//
// export interface TsTypeFnOrConstructorDef extends TsTypeDefBase {
//   kind: "fnOrConstructor";
//   fnOrConstructor: TsFnOrConstructorDef;
// }
//
// export interface TsTypeConditionalDef extends TsTypeDefBase {
//   kind: "conditional";
//   conditionalType: TsConditionalDef;
// }
//
// export interface TsTypeInferDef extends TsTypeDefBase {
//   kind: "infer";
//   infer: TsInferDef;
// }
//
// export interface TsTypeMappedDef extends TsTypeDefBase {
//   kind: "mapped";
//   mappedType: TsMappedTypeDef;
// }
//
// export interface TsTypeImportTypeDef extends TsTypeDefBase {
//   kind: "importType";
//   importType: TsImportTypeDef;
// }
//
// export interface TsTypeIndexedAccessDef extends TsTypeDefBase {
//   kind: "indexedAccess";
//   indexedAccess: TsIndexedAccessDef;
// }
//
// export interface TsTypeTypeLiteralDef extends TsTypeDefBase {
//   kind: "typeLiteral";
//   typeLiteral: TsTypeLiteralDef;
// }
//
// export interface TsTypeTypePredicateDef extends TsTypeDefBase {
//   kind: "typePredicate";
//   typePredicate: TsTypePredicateDef;
// }
//
// export type TsTypeDefKind =
//   | "keyword"
//   | "literal"
//   | "typeRef"
//   | "union"
//   | "intersection"
//   | "array"
//   | "tuple"
//   | "typeOperator"
//   | "parenthesized"
//   | "rest"
//   | "optional"
//   | "typeQuery"
//   | "this"
//   | "fnOrConstructor"
//   | "conditional"
//   | "importType"
//   | "infer"
//   | "indexedAccess"
//   | "mapped"
//   | "typeLiteral"
//   | "typePredicate";
//
// export interface TsTypeRefDef {
//   typeParams?: TsTypeDef[];
//   typeName: string;
// }
//
// export interface TypeAliasDef {
//   tsType: TsTypeDef;
//   typeParams: TsTypeParamDef[];
// }
//
// export type VariableDeclKind = "var" | "let" | "const";
//
// export interface VariableDef {
//   tsType?: TsTypeDef;
//   kind: VariableDeclKind;
// }

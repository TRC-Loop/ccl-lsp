package main

import (
	"github.com/TRC-Loop/ccolon/parser"
)

type SymbolKind int

const (
	KindVariable SymbolKind = iota
	KindConst
	KindFunction
	KindClass
	KindParameter
	KindField
	KindMethod
	KindModule
)

type ParamInfo struct {
	Name     string
	TypeName string
	Optional bool
}

type FieldInfo struct {
	Name       string
	TypeName   string
	Visibility string
}

type MethodInfo struct {
	Name       string
	Params     []ParamInfo
	ReturnType string
	Visibility string
}

type Symbol struct {
	Name       string
	Kind       SymbolKind
	TypeName   string
	Line       int
	Col        int
	URI        string
	Params     []ParamInfo
	ReturnType string
	Fields     []FieldInfo
	Methods    []MethodInfo
	SuperName  string
}

type Analysis struct {
	Globals map[string]*Symbol
}

var stdlibModules = map[string]*Symbol{
	"console":  {Name: "console", Kind: KindModule, TypeName: "module"},
	"math":     {Name: "math", Kind: KindModule, TypeName: "module"},
	"random":   {Name: "random", Kind: KindModule, TypeName: "module"},
	"json":     {Name: "json", Kind: KindModule, TypeName: "module"},
	"fs":       {Name: "fs", Kind: KindModule, TypeName: "module"},
	"datetime": {Name: "datetime", Kind: KindModule, TypeName: "module"},
	"os":       {Name: "os", Kind: KindModule, TypeName: "module"},
	"http":     {Name: "http", Kind: KindModule, TypeName: "module"},
}

var builtinMethods = map[string][]MethodInfo{
	"string": {
		{Name: "length", ReturnType: "int"},
		{Name: "tostring", ReturnType: "string"},
		{Name: "toint", ReturnType: "int"},
		{Name: "tofloat", ReturnType: "float"},
		{Name: "tosint", ReturnType: "sint"},
		{Name: "split", Params: []ParamInfo{{Name: "sep", TypeName: "string"}}, ReturnType: "list"},
		{Name: "reverse", ReturnType: "string"},
		{Name: "upper", ReturnType: "string"},
		{Name: "lower", ReturnType: "string"},
		{Name: "trim", ReturnType: "string"},
		{Name: "contains", Params: []ParamInfo{{Name: "sub", TypeName: "string"}}, ReturnType: "bool"},
		{Name: "startswith", Params: []ParamInfo{{Name: "prefix", TypeName: "string"}}, ReturnType: "bool"},
		{Name: "endswith", Params: []ParamInfo{{Name: "suffix", TypeName: "string"}}, ReturnType: "bool"},
		{Name: "replace", Params: []ParamInfo{{Name: "old", TypeName: "string"}, {Name: "new", TypeName: "string"}}, ReturnType: "string"},
		{Name: "index", Params: []ParamInfo{{Name: "sub", TypeName: "string"}}, ReturnType: "int"},
		{Name: "repeat", Params: []ParamInfo{{Name: "n", TypeName: "int"}}, ReturnType: "string"},
		{Name: "join", Params: []ParamInfo{{Name: "list", TypeName: "list"}}, ReturnType: "string"},
	},
	"int": {
		{Name: "tostring", ReturnType: "string"},
		{Name: "tofloat", ReturnType: "float"},
		{Name: "tosint", ReturnType: "sint"},
		{Name: "abs", ReturnType: "int"},
		{Name: "pow", Params: []ParamInfo{{Name: "exp", TypeName: "int"}}, ReturnType: "int"},
	},
	"float": {
		{Name: "tostring", ReturnType: "string"},
		{Name: "toint", ReturnType: "int"},
		{Name: "abs", ReturnType: "float"},
		{Name: "pow", Params: []ParamInfo{{Name: "exp", TypeName: "float"}}, ReturnType: "float"},
	},
	"sint": {
		{Name: "tostring", ReturnType: "string"},
		{Name: "toint", ReturnType: "int"},
		{Name: "tofloat", ReturnType: "float"},
		{Name: "abs", ReturnType: "sint"},
		{Name: "pow", Params: []ParamInfo{{Name: "exp", TypeName: "sint"}}, ReturnType: "sint"},
	},
	"bool": {
		{Name: "tostring", ReturnType: "string"},
	},
	"list": {
		{Name: "length", ReturnType: "int"},
		{Name: "append", Params: []ParamInfo{{Name: "value", TypeName: "any"}}, ReturnType: ""},
		{Name: "pop", ReturnType: "any"},
		{Name: "tostring", ReturnType: "string"},
	},
	"array": {
		{Name: "length", ReturnType: "int"},
		{Name: "tostring", ReturnType: "string"},
	},
	"dict": {
		{Name: "keys", ReturnType: "list"},
		{Name: "values", ReturnType: "list"},
		{Name: "has", Params: []ParamInfo{{Name: "key", TypeName: "string"}}, ReturnType: "bool"},
		{Name: "length", ReturnType: "int"},
		{Name: "tostring", ReturnType: "string"},
	},
}

var stdlibMethods = map[string][]MethodInfo{
	"console": {
		{Name: "println", Params: []ParamInfo{{Name: "value", TypeName: "any"}}},
		{Name: "print", Params: []ParamInfo{{Name: "value", TypeName: "any"}}},
		{Name: "readLine", ReturnType: "string"},
		{Name: "readInt", ReturnType: "int"},
	},
	"math": {
		{Name: "sqrt", Params: []ParamInfo{{Name: "x", TypeName: "float"}}, ReturnType: "float"},
		{Name: "abs", Params: []ParamInfo{{Name: "x", TypeName: "float"}}, ReturnType: "float"},
		{Name: "floor", Params: []ParamInfo{{Name: "x", TypeName: "float"}}, ReturnType: "int"},
		{Name: "ceil", Params: []ParamInfo{{Name: "x", TypeName: "float"}}, ReturnType: "int"},
		{Name: "pow", Params: []ParamInfo{{Name: "base", TypeName: "float"}, {Name: "exp", TypeName: "float"}}, ReturnType: "float"},
		{Name: "min", Params: []ParamInfo{{Name: "a", TypeName: "float"}, {Name: "b", TypeName: "float"}}, ReturnType: "float"},
		{Name: "max", Params: []ParamInfo{{Name: "a", TypeName: "float"}, {Name: "b", TypeName: "float"}}, ReturnType: "float"},
		{Name: "pi", ReturnType: "float"},
		{Name: "e", ReturnType: "float"},
		{Name: "sin", Params: []ParamInfo{{Name: "x", TypeName: "float"}}, ReturnType: "float"},
		{Name: "cos", Params: []ParamInfo{{Name: "x", TypeName: "float"}}, ReturnType: "float"},
		{Name: "tan", Params: []ParamInfo{{Name: "x", TypeName: "float"}}, ReturnType: "float"},
	},
	"random": {
		{Name: "randint", Params: []ParamInfo{{Name: "min", TypeName: "int"}, {Name: "max", TypeName: "int"}}, ReturnType: "int"},
		{Name: "randfloat", Params: []ParamInfo{{Name: "min", TypeName: "float"}, {Name: "max", TypeName: "float"}}, ReturnType: "float"},
		{Name: "choice", Params: []ParamInfo{{Name: "list", TypeName: "list"}}, ReturnType: "any"},
		{Name: "char", ReturnType: "string"},
	},
	"json": {
		{Name: "parse", Params: []ParamInfo{{Name: "text", TypeName: "string"}}, ReturnType: "dict"},
		{Name: "stringify", Params: []ParamInfo{{Name: "value", TypeName: "any"}}, ReturnType: "string"},
	},
	"fs": {
		{Name: "open", Params: []ParamInfo{{Name: "path", TypeName: "string"}, {Name: "mode", TypeName: "string"}}, ReturnType: "file"},
		{Name: "read", Params: []ParamInfo{{Name: "path", TypeName: "string"}}, ReturnType: "string"},
		{Name: "write", Params: []ParamInfo{{Name: "path", TypeName: "string"}, {Name: "content", TypeName: "string"}}},
		{Name: "exists", Params: []ParamInfo{{Name: "path", TypeName: "string"}}, ReturnType: "bool"},
		{Name: "delete", Params: []ParamInfo{{Name: "path", TypeName: "string"}}},
	},
	"datetime": {
		{Name: "now", ReturnType: "string"},
		{Name: "parse", Params: []ParamInfo{{Name: "text", TypeName: "string"}}, ReturnType: "dict"},
		{Name: "format", Params: []ParamInfo{{Name: "dt", TypeName: "dict"}, {Name: "fmt", TypeName: "string"}}, ReturnType: "string"},
	},
	"os": {
		{Name: "getenv", Params: []ParamInfo{{Name: "key", TypeName: "string"}}, ReturnType: "string"},
		{Name: "setenv", Params: []ParamInfo{{Name: "key", TypeName: "string"}, {Name: "value", TypeName: "string"}}},
		{Name: "exit", Params: []ParamInfo{{Name: "code", TypeName: "int"}}},
		{Name: "system", Params: []ParamInfo{{Name: "cmd", TypeName: "string"}}, ReturnType: "string"},
	},
	"http": {
		{Name: "get", Params: []ParamInfo{{Name: "url", TypeName: "string"}}, ReturnType: "string"},
		{Name: "post", Params: []ParamInfo{{Name: "url", TypeName: "string"}, {Name: "body", TypeName: "string"}}, ReturnType: "string"},
		{Name: "put", Params: []ParamInfo{{Name: "url", TypeName: "string"}, {Name: "body", TypeName: "string"}}, ReturnType: "string"},
		{Name: "delete", Params: []ParamInfo{{Name: "url", TypeName: "string"}}, ReturnType: "string"},
	},
}

func analyzeAST(uri string, prog *parser.Program) *Analysis {
	a := &Analysis{Globals: make(map[string]*Symbol)}

	for name, sym := range stdlibModules {
		clone := *sym
		a.Globals[name] = &clone
	}

	if prog == nil {
		return a
	}

	for _, stmt := range prog.Stmts {
		switch s := stmt.(type) {
		case *parser.ImportStmt:
			name := s.Alias
			if name == "" {
				name = s.Module
			}
			a.Globals[name] = &Symbol{
				Name:     name,
				Kind:     KindModule,
				TypeName: "module",
				Line:     s.Pos().Line,
				Col:      s.Pos().Col,
				URI:      uri,
			}

		case *parser.FromImportStmt:
			for _, name := range s.Names {
				if name == "*" {
					continue
				}
				a.Globals[name] = &Symbol{
					Name: name,
					Kind: KindModule,
					Line: s.Pos().Line,
					URI:  uri,
				}
			}

		case *parser.VarDecl:
			kind := KindVariable
			if s.IsConst {
				kind = KindConst
			}
			a.Globals[s.Name] = &Symbol{
				Name:     s.Name,
				Kind:     kind,
				TypeName: s.TypeName,
				Line:     s.Pos().Line,
				Col:      s.Pos().Col,
				URI:      uri,
			}

		case *parser.FuncDecl:
			sym := &Symbol{
				Name:       s.Name,
				Kind:       KindFunction,
				ReturnType: s.ReturnType,
				Line:       s.Pos().Line,
				Col:        s.Pos().Col,
				URI:        uri,
			}
			for _, p := range s.Params {
				sym.Params = append(sym.Params, ParamInfo{
					Name:     p.Name,
					TypeName: p.TypeName,
					Optional: p.Default != nil,
				})
			}
			a.Globals[s.Name] = sym

		case *parser.ClassDecl:
			sym := &Symbol{
				Name:      s.Name,
				Kind:      KindClass,
				SuperName: s.SuperName,
				Line:      s.Pos().Line,
				Col:       s.Pos().Col,
				URI:       uri,
			}
			for _, f := range s.Fields {
				sym.Fields = append(sym.Fields, FieldInfo{
					Name:       f.Name,
					TypeName:   f.TypeName,
					Visibility: f.Visibility,
				})
			}
			for _, m := range s.Methods {
				mi := MethodInfo{
					Name:       m.Name,
					ReturnType: m.ReturnType,
					Visibility: m.Visibility,
				}
				for _, p := range m.Params {
					mi.Params = append(mi.Params, ParamInfo{
						Name:     p.Name,
						TypeName: p.TypeName,
						Optional: p.Default != nil,
					})
				}
				sym.Methods = append(sym.Methods, mi)
			}
			a.Globals[s.Name] = sym
		}
	}

	return a
}

// resolveMembers returns the members of a symbol for dot-completion.
func (a *Analysis) resolveMembers(name string) []MethodInfo {
	sym, ok := a.Globals[name]
	if !ok {
		return nil
	}
	switch sym.Kind {
	case KindModule:
		if methods, ok := stdlibMethods[name]; ok {
			return methods
		}
	case KindClass:
		var out []MethodInfo
		for _, f := range sym.Fields {
			out = append(out, MethodInfo{Name: f.Name, ReturnType: f.TypeName})
		}
		out = append(out, sym.Methods...)
		if sym.SuperName != "" {
			out = append(out, a.resolveMembers(sym.SuperName)...)
		}
		return out
	}
	if methods, ok := builtinMethods[sym.TypeName]; ok {
		return methods
	}
	return nil
}

// formatSignature returns a human-readable signature for a symbol.
func formatSignature(sym *Symbol) string {
	switch sym.Kind {
	case KindFunction:
		return sprintf("function %s(%s) %s", sym.Name, formatParams(sym.Params), sym.ReturnType)
	case KindClass:
		base := sprintf("class %s", sym.Name)
		if sym.SuperName != "" {
			base += " extends " + sym.SuperName
		}
		return base
	case KindVariable:
		return sprintf("var %s %s", sym.TypeName, sym.Name)
	case KindConst:
		return sprintf("const %s %s", sym.TypeName, sym.Name)
	case KindModule:
		return sprintf("module %s", sym.Name)
	}
	return sym.Name
}

func formatParams(params []ParamInfo) string {
	parts := make([]string, len(params))
	for i, p := range params {
		s := p.TypeName + " " + p.Name
		if p.Optional {
			s += "?"
		}
		parts[i] = s
	}
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}

func formatMethodSignature(m MethodInfo) string {
	sig := sprintf("function %s(%s)", m.Name, formatParams(m.Params))
	if m.ReturnType != "" {
		sig += " " + m.ReturnType
	}
	return sig
}

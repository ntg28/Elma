package gen

import (
    "fmt"
    "go/ast"
    "strings"
    "reflect"
    "golang.org/x/tools/go/packages"
)

type binds map[string]string

type Gen struct {
    Pkgs []*packages.Package
    Binds binds
    // depth:
    //   If depth is bigger than 0 we are inside a `(...)` or `[...]`
    //   otherwise we are inside a function this is usefull for knowing
    //   where to add the `;` symbol.
    depth int
}

func (gen *Gen) AddDepth() {
    gen.depth++
}

func (gen *Gen) RemDepth() {
    gen.depth--
}

func (gen *Gen) AddSemicolon() bool {
    return gen.depth == 0
}

func (gen *Gen) Decls() []ast.Decl {
    var out []ast.Decl
    for _, pkg := range gen.Pkgs {
        for _, file := range pkg.Syntax {
            for _, decl := range file.Decls {
                out = append(out, decl)
            }
        }
    }
    return out
}

func (gen *Gen) LookupType(name string) *ast.TypeSpec {
    for _, decl := range gen.Decls() {
        switch e := decl.(type) {
            case *ast.GenDecl: {
                if e.Tok.String() == "type" {
                    for _, type_ := range e.Specs {
                        t, _ := type_.(*ast.TypeSpec)
                        if t.Name.Name == name {
                            return t
                        }
                    }
                }
            }
            default: {}
        }
    }
    return nil
}

func (gen *Gen) LookupFunc(name string) *ast.FuncDecl {
    for _, decl := range gen.Decls() {
        switch e := decl.(type) {
            case *ast.FuncDecl: {
                if e.Name.Name == name {
                    return e
                }
            }
            default: {}
        }
    }
    return nil
}

func (gen *Gen) GenPkgs() string {
    var out string
    for _, pkg := range gen.Pkgs {
        out += gen.GenPkg(pkg)
    }
    return out
}

func (gen *Gen) GenPkg(pkg *packages.Package) string {
    var out string
    for _, file := range pkg.Syntax {
        out += gen.GenFile(file)
    }
    return out
}

func (gen *Gen) GenFile(file *ast.File) string {
    var out string
    for _, decl := range file.Decls {
        switch e := decl.(type) {
            case *ast.FuncDecl: {
                out += gen.GenFuncDecl(e)
            }
            case *ast.GenDecl: {
                if e.Tok.String() == "import" {
                    continue
                }
                out += gen.GenGenDecl(e)
            }
            default: {}
        }
    }
    return out
}

func (gen *Gen) GenFuncDecl(fun *ast.FuncDecl) string {
    var out string

    if fun.Recv != nil && len(fun.Recv.List) != 1 {
        panic(fmt.Sprintf("GenFuncDecl cannot generate function '%s'", fun.Name.Name))
    }

    if fun.Recv != nil {
        out += gen.GenExpr(fun.Recv.List[0].Type)
        out += ".prototype."
        out += fun.Name.Name
        out += "="
    }

    out += "function "

    if fun.Recv == nil {
        out += fun.Name.Name
    }

    out += "("
    out += gen.GenFields(fun.Type.Params)
    out += ")"
    out += "{"

    if fun.Recv != nil {
        gen.Binds[gen.GenIdent(fun.Recv.List[0].Names[0])] = "this"
    }

    out += gen.GenBlockStmt(fun.Body)

    gen.Binds = map[string]string{}

    out += "}"

    if fun.Recv != nil {
        out += ";"
    }
    return out
}

func (gen *Gen) GenStmt(stmt ast.Stmt) string {
    switch t := stmt.(type) {
        case *ast.IfStmt: return gen.GenIfStmt(t)
        case *ast.ForStmt: return gen.GenForStmt(t)
        case *ast.DeclStmt: return gen.GenDeclStmt(t)
        case *ast.ExprStmt: return gen.GenExprStmt(t)
        case *ast.ReturnStmt: return gen.GenReturnStmt(t)
        case *ast.AssignStmt: return gen.GenAssignStmt(t)
        case *ast.IncDecStmt: return gen.GenIncDecStmt(t)
        case *ast.SwitchStmt: return gen.GenSwitchStmt(t)
        case *ast.CaseClause: return gen.GenCaseClause(t)
        case *ast.RangeStmt: return gen.GenRangeStmt(t)
        case *ast.BlockStmt: return gen.GenBlockStmt(t)
        case nil: return ""
        default: {
            panic(fmt.Sprintf("GenStmt not implemented for (%+v)", reflect.TypeOf(stmt)))
        }
    }
}

func (gen *Gen) GenReturnStmt(stmt *ast.ReturnStmt) string {
    var out string
    out += "return "

    gen.AddDepth()

    for _, expr := range stmt.Results {
        out += gen.GenExpr(expr)
    }

    gen.RemDepth()

    out += ";"
    return out
}

func (gen *Gen) GenBlockStmt(expr *ast.BlockStmt) string {
    var out string
    for _, stmt := range expr.List {
        out += gen.GenStmt(stmt)
    }
    return out
}

func (gen *Gen) GenIfStmt(expr *ast.IfStmt) string {
    gen.AddDepth()

    var cond string = gen.GenExpr(expr.Cond)

    gen.RemDepth()

    var body string = gen.GenBlockStmt(expr.Body)
    var elsi string = ""

    if expr.Else != nil {
        elsi += "else{" + gen.GenStmt(expr.Else) + "}"
    }

    return "if (" + cond + ")" + "{" + body + "}" + elsi
}

func (gen *Gen) GenForStmt(expr *ast.ForStmt) string {
    gen.AddDepth()

    var init string = gen.GenStmt(expr.Init)
    var cond string = gen.GenExpr(expr.Cond)
    var post string = gen.GenStmt(expr.Post)

    gen.RemDepth()
    
    var body string = gen.GenBlockStmt(expr.Body)

    return "for (" + init + ";" + cond + ";" + post + ")" + "{" + body + "}"
}

func (gen *Gen) GenIncDecStmt(expr *ast.IncDecStmt) string {
    out := gen.GenExpr(expr.X) + expr.Tok.String()
    if gen.AddSemicolon() {
        out += ";"
    }
    return out
}

func (gen *Gen) GenCaseClause(expr *ast.CaseClause) string {
    out := ""
    if len(expr.List) < 1 {
        out += "default"
    } else {
        out += "case " + gen.GenExpr(expr.List[0])
    }
    out += ":{"
    for _, stmt := range expr.Body {
        out += gen.GenStmt(stmt)
    }
    out += "}"
    return out
}

func (gen *Gen) GenSwitchStmt(expr *ast.SwitchStmt) string {
    tag := gen.GenExpr(expr.Tag)
    body := gen.GenBlockStmt(expr.Body)
    return "switch" + tag + "{" + body + "}"
}

func (gen *Gen) GenExpr(expr ast.Expr) string {
    switch e := expr.(type) {
        case *ast.Ident: return gen.GenIdent(e)
        case *ast.BasicLit: return gen.GenBasicLit(e)
        case *ast.CallExpr: return gen.GenCall(e)
        case *ast.BinaryExpr: return gen.GenBinaryExpr(e)
        case *ast.SelectorExpr: return gen.GenSelector(e)
        case *ast.CompositeLit: return gen.GenCompositeLit(e)
        case *ast.KeyValueExpr: return gen.GenKeyValueExpr(e)
        case *ast.ArrayType: return gen.GenArrayType(e)
        case *ast.UnaryExpr: return gen.GenUnaryExpr(e)
        case *ast.ParenExpr: return gen.GenParenExpr(e)
        case *ast.FuncLit: return gen.GenFuncLit(e)
        default: {
            panic(fmt.Sprintf("GenExpr not implemented for (%+v)", reflect.TypeOf(expr)))
        }
    }
}

func (gen *Gen) GenIdent(expr *ast.Ident) string {
    if val, ok := gen.Binds[expr.Name]; ok {
        return val
    } else {
        return expr.Name
    }
}

func (gen *Gen) GenBinaryExpr(expr *ast.BinaryExpr) string {
    return gen.GenExpr(expr.X) + expr.Op.String() + gen.GenExpr(expr.Y)
}

func (gen *Gen) GenBasicLit(expr *ast.BasicLit) string {
    return expr.Value
}

func (gen *Gen) GenGenDecl(expr *ast.GenDecl) string {
    var out string
    for _, spec := range expr.Specs {
        out += gen.GenSpec(spec)
    }
    return out
}

func (gen *Gen) GenDeclStmt(expr *ast.DeclStmt) string {
    a, isGenDecl := expr.Decl.(*ast.GenDecl)
    if isGenDecl {
        return gen.GenGenDecl(a)
    }
    panic(fmt.Sprintf("GenDecl not implemented for (%v)", reflect.TypeOf(expr.Decl)))
}

func (gen *Gen) GenUnaryExpr(expr *ast.UnaryExpr) string {
    return expr.Op.String() + gen.GenExpr(expr.X)
}

func (gen *Gen) GenSpec(expr ast.Spec) string {
    switch e := expr.(type) {
        case *ast.TypeSpec: return gen.GenTypeSpec(e)
        case *ast.ValueSpec: return gen.GenValueSpec(e)
        default: { 
            panic(fmt.Sprintf("GenSpec not implemented for (%v)", reflect.TypeOf(expr)))
        }
    }
}

func (gen *Gen) GenValueSpec(expr *ast.ValueSpec) string {
    var out string
    for i, name := range expr.Names {
        out += "let "
        out += gen.GenIdent(name)
        if i < len(expr.Values) {
            out += "="
            out += gen.GenExpr(expr.Values[i])
        }
        out += ";"
    }
    return out
}

func (gen *Gen) GenAssignStmt(expr *ast.AssignStmt) string {
    if len(expr.Lhs) < 1 { panic("missing lhs of assignment") }
    if len(expr.Rhs) < 1 { panic("missing rhs of assignment") }

    tok := expr.Tok.String()
    if tok == ":=" {
        tok = "="
    }

    var out string
    out += gen.GenExpr(expr.Lhs[0])
    out += tok

    gen.AddDepth()

    out += gen.GenExpr(expr.Rhs[0])

    gen.RemDepth()

    if gen.AddSemicolon() {
        out += ";"
    }
    return out
}

func isJsBindFunc(expr *ast.FuncDecl) bool {
    return expr != nil                  &&
           expr.Doc != nil              &&
           len(expr.Doc.List) > 1       &&
           expr.Doc.List[0].Text[2:] == "js-bind"
}

func (gen *Gen) GenCall(expr *ast.CallExpr) string {
    name := gen.GenExpr(expr.Fun)
    sels := strings.Split(name, ".")
    query := ""
    out := ""

    if _, ok := expr.Fun.(*ast.SelectorExpr); ok {
        query = sels[len(sels) - 1]
    } else {
        query = name
    }

    fun := gen.LookupFunc(query)

    gen.AddDepth()

    var args string
    for i, arg := range expr.Args {
        args += gen.GenExpr(arg)
        if i < len(expr.Args) - 1 {
            args += ","
        }
    }

    gen.RemDepth()

    if !isJsBindFunc(fun) {
        out = name + "(" + args + ")"
    } else {
        for _, doc := range fun.Doc.List[1:] {
            out += doc.Text[2:]
        }
        out = strings.Replace(out, "%args%", args, 1)
        out = strings.Replace(out, "%recv%", strings.Join(sels[:len(sels)-1], "."), 1)
        for i, arg := range expr.Args {
            out = strings.Replace(out, fmt.Sprintf("%%arg%d%%", i), gen.GenExpr(arg), -1)
        }
    }

    if gen.AddSemicolon() {
        out += ";"
    }
    return out
}

func (gen *Gen) GenField(expr *ast.Field) string {
    if len(expr.Names) < 1 { panic("missing field name") }
    return gen.GenIdent(expr.Names[0])
}

func (gen *Gen) GenExprStmt(expr *ast.ExprStmt) string {
    return gen.GenExpr(expr.X)
}

func (gen *Gen) GenSelector(expr *ast.SelectorExpr) string {
    parent := gen.GenExpr(expr.X)
    callee := gen.GenExpr(expr.Sel)
    return parent + "." + callee
}

func (gen *Gen) GenKeyValueExpr(expr *ast.KeyValueExpr) string {
    key := gen.GenExpr(expr.Key)
    val := gen.GenExpr(expr.Value)
    return key + ":" + val
}

func (gen *Gen) GenFields(fields *ast.FieldList) string {
    var out string
    for i, param := range fields.List {
        out += gen.GenField(param)
        if i < len(fields.List) - 1 {
            out += ","
        }
    }
    return out
}

func (gen *Gen) GenTypeSpec(expr *ast.TypeSpec) string {
    var out string
    a, isStruct := expr.Type.(*ast.StructType)
    if isStruct {
        out += "function "
        out += expr.Name.Name
        out += "("
        if a.Fields.List != nil {
            out += gen.GenFields(a.Fields)
        }
        out += ")"
        out += "{"
        for _, field := range a.Fields.List {
            field_str := gen.GenField(field)
            out += "this." + field_str + "=" + field_str + ";"
        }
        out += "}"
        return out
    }
    panic(fmt.Sprintf("GenTypeSpec not implemented for type (%v)", reflect.TypeOf(expr.Type)))
}

func (gen *Gen) LookupCompositeType(expr *ast.CompositeLit) *ast.TypeSpec {
    return gen.LookupType(gen.GenExpr(expr.Type))
}

func (gen *Gen) GenStructConstructor(expr *ast.CompositeLit) string {
    fields := ""
    type_str := gen.GenExpr(expr.Type)

    for i, field := range expr.Elts {
        switch e := field.(type) {
            case *ast.KeyValueExpr: {
                fields += gen.GenExpr(e.Value)
                if i < len(expr.Elts) - 1 {
                    fields += ","
                }
            }
            case *ast.BasicLit: {
                fields += e.Value
                if i < len(expr.Elts) - 1 {
                    fields += ","
                }
            }
            default: {}
        }
    }
    return "new " + type_str + "(" + fields + ")"
}

func (gen *Gen) GenCompositeLit(expr *ast.CompositeLit) string {
    type_decl := gen.LookupCompositeType(expr)
    if type_decl != nil {
        return gen.GenStructConstructor(expr);
    } else {
        left := "{"
        right := "}"
        fields := ""

        if _, ok := expr.Type.(*ast.ArrayType); ok {
            left = "["
            right = "]"
        }

        for i, field := range expr.Elts {
            fields += gen.GenExpr(field)
            if i < len(expr.Elts) - 1 {
                fields += ","
            }
        }

        return left + fields + right
    }
}

func (gen *Gen) GenArrayType(expr *ast.ArrayType) string {
    return "???"
}

func (gen *Gen) GenRangeStmt(expr *ast.RangeStmt) string {
    var key string = gen.GenExpr(expr.Key)
    var val string = gen.GenExpr(expr.Value)
    var subj string = gen.GenExpr(expr.X)
    var body string = gen.GenBlockStmt(expr.Body)

    return fmt.Sprintf(
        "for (let [%s, %s] of Object.entries(%s)) {%s}",
        key, val, subj, body,
    )
}

func (gen *Gen) GenParenExpr(expr *ast.ParenExpr) string {
    return "(" + gen.GenExpr(expr.X) + ")"
}

func (gen *Gen) GenFuncLit(expr *ast.FuncLit) string {
    var out string
    out += "() => {"
    out += gen.GenBlockStmt(expr.Body)
    out += "}"
    return out
}

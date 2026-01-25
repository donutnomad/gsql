// Command export-types generates type aliases and constructor wrappers
// from internal packages to a public-facing Go file.
//
// Usage:
//
//	go run ./cmd/export-types -src ./internal/fields -dst ./typed_field.go -pkg gsql
//	go run ./cmd/export-types -src ./internal/fields -dst ./typed_field.go -pkg gsql -exclude "helper,internal"
//	go run ./cmd/export-types -src ./internal/clauses/casewhen.go -dst ./clause_case.go -pkg gsql
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/donutnomad/gsql/internal/genutils"
)

type Config struct {
	SrcDir      string
	DstFile     string
	PkgName     string
	ExcludeList []string
	GenCmd      string // go:generate 命令
}

// TypeInfo 存储类型信息
type TypeInfo struct {
	Name       string
	IsGeneric  bool
	TypeParams string // 泛型参数定义，如 "[T any]"
	TypeArgs   string // 泛型类型参数，如 "[T]"
	SrcPkg     string // 源包名
}

// FuncInfo 存储函数信息
type FuncInfo struct {
	Name           string
	IsGeneric      bool
	TypeParams     string   // 泛型参数定义，如 "[T any]"
	TypeParamNames []string // 泛型参数名称列表，如 ["T"]
	Params         string   // 函数参数
	Results        string   // 返回值
	SrcPkg         string   // 源包名
	CallArgs       string   // 调用参数
	TypeArgs       string   // 泛型类型参数，如 "[T]"
	Doc            string   // 文档注释
}

// VarInfo 存储变量信息
type VarInfo struct {
	Name   string
	SrcPkg string // 源包名
	Doc    string // 文档注释
}

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	var cfg Config
	var excludeStr string

	flag.StringVar(&cfg.SrcDir, "src", "", "Source file or directory containing Go files")
	flag.StringVar(&cfg.DstFile, "dst", "", "Destination file for generated code")
	flag.StringVar(&cfg.PkgName, "pkg", "", "Package name for generated file")
	flag.StringVar(&excludeStr, "exclude", "", "Comma-separated list of names to exclude")
	flag.StringVar(&cfg.GenCmd, "gencmd", "", "go:generate command to embed in output (optional)")
	flag.Parse()

	if cfg.SrcDir == "" || cfg.DstFile == "" || cfg.PkgName == "" {
		fmt.Println("Usage: export-types -src <file-or-dir> -dst <file> -pkg <package> [-exclude <names>]")
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if excludeStr != "" {
		cfg.ExcludeList = strings.Split(excludeStr, ",")
		for i := range cfg.ExcludeList {
			cfg.ExcludeList[i] = strings.TrimSpace(cfg.ExcludeList[i])
		}
	}

	return cfg
}

func run(cfg Config) error {
	fset := token.NewFileSet()

	// 收集类型、函数和变量
	var types []TypeInfo
	var funcs []FuncInfo
	var vars []VarInfo
	var srcPkgName string
	var srcPath string                 // 用于计算导入路径的路径（目录）
	imports := make(map[string]string) // path -> alias

	// 检查源路径是文件还是目录
	info, err := os.Stat(cfg.SrcDir)
	if err != nil {
		return fmt.Errorf("failed to stat source path: %w", err)
	}

	if info.IsDir() {
		// 解析目录
		srcPath = cfg.SrcDir
		pkgs, err := parser.ParseDir(fset, cfg.SrcDir, func(fi os.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse source directory: %w", err)
		}

		for pkgName, pkg := range pkgs {
			// 跳过 main 包（通常是代码生成器）
			if pkgName == "main" {
				continue
			}
			srcPkgName = pkgName
			for _, file := range pkg.Files {
				collectImports(file, imports)
				types = append(types, collectTypes(file, pkgName, cfg.ExcludeList)...)
				funcs = append(funcs, collectFuncs(file, pkgName, cfg.ExcludeList)...)
				vars = append(vars, collectVars(file, pkgName, cfg.ExcludeList)...)
			}
		}
	} else {
		// 解析单个文件
		srcPath = filepath.Dir(cfg.SrcDir)
		file, err := parser.ParseFile(fset, cfg.SrcDir, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse source file: %w", err)
		}

		srcPkgName = file.Name.Name
		collectImports(file, imports)
		types = append(types, collectTypes(file, srcPkgName, cfg.ExcludeList)...)
		funcs = append(funcs, collectFuncs(file, srcPkgName, cfg.ExcludeList)...)
		vars = append(vars, collectVars(file, srcPkgName, cfg.ExcludeList)...)
	}

	// 排序以确保稳定输出
	sort.Slice(types, func(i, j int) bool { return types[i].Name < types[j].Name })
	// 函数排序：New 开头的放前面，其余按字母顺序
	sort.Slice(funcs, func(i, j int) bool {
		iIsNew := strings.HasPrefix(funcs[i].Name, "New")
		jIsNew := strings.HasPrefix(funcs[j].Name, "New")
		if iIsNew != jIsNew {
			return iIsNew // New 开头的排前面
		}
		return funcs[i].Name < funcs[j].Name
	})
	sort.Slice(vars, func(i, j int) bool { return vars[i].Name < vars[j].Name })

	// 计算导入路径
	srcImportPath, err := getImportPath(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get import path: %w", err)
	}

	// 生成代码
	code := generateCode(cfg, srcPkgName, srcImportPath, types, funcs, vars)

	// 写入文件
	if err := genutils.WriteFormat(cfg.DstFile, []byte(code)); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Generated %s with %d types, %d functions and %d variables\n", cfg.DstFile, len(types), len(funcs), len(vars))
	return nil
}

func collectImports(file *ast.File, imports map[string]string) {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		imports[path] = alias
	}
}

func collectTypes(file *ast.File, pkgName string, excludeList []string) []TypeInfo {
	var types []TypeInfo

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			name := typeSpec.Name.Name
			// 只导出公共类型
			if !ast.IsExported(name) {
				continue
			}

			// 检查排除列表
			if isExcluded(name, excludeList) {
				continue
			}

			info := TypeInfo{
				Name:   name,
				SrcPkg: pkgName,
			}

			// 检查是否有泛型参数
			if typeSpec.TypeParams != nil && len(typeSpec.TypeParams.List) > 0 {
				info.IsGeneric = true
				info.TypeParams = formatTypeParams(typeSpec.TypeParams)
				info.TypeArgs = formatTypeArgs(typeSpec.TypeParams)
			}

			types = append(types, info)
		}
	}

	return types
}

func collectFuncs(file *ast.File, pkgName string, excludeList []string) []FuncInfo {
	var funcs []FuncInfo

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// 只处理包级函数（不是方法）
		if funcDecl.Recv != nil {
			continue
		}

		name := funcDecl.Name.Name

		// 只导出公共函数
		if !ast.IsExported(name) {
			continue
		}

		// 检查排除列表
		if isExcluded(name, excludeList) {
			continue
		}

		info := FuncInfo{
			Name:   name,
			SrcPkg: pkgName,
		}

		// 收集文档注释
		if funcDecl.Doc != nil {
			info.Doc = funcDecl.Doc.Text()
		}

		// 检查泛型参数
		if funcDecl.Type.TypeParams != nil && len(funcDecl.Type.TypeParams.List) > 0 {
			info.IsGeneric = true
			info.TypeParams = formatTypeParams(funcDecl.Type.TypeParams)
			info.TypeArgs = formatTypeArgs(funcDecl.Type.TypeParams)
			info.TypeParamNames = collectTypeParamNames(funcDecl.Type.TypeParams)
		}

		// 格式化参数
		info.Params = formatParams(funcDecl.Type.Params)
		info.CallArgs = formatCallArgs(funcDecl.Type.Params)

		// 格式化返回值（排除泛型参数名称）
		if funcDecl.Type.Results != nil {
			info.Results = formatResultsWithExcludes(funcDecl.Type.Results, pkgName, info.TypeParamNames)
		}

		funcs = append(funcs, info)
	}

	return funcs
}

func collectVars(file *ast.File, pkgName string, excludeList []string) []VarInfo {
	var vars []VarInfo

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, name := range valueSpec.Names {
				// 只导出公共变量
				if !ast.IsExported(name.Name) {
					continue
				}

				// 检查排除列表
				if isExcluded(name.Name, excludeList) {
					continue
				}

				info := VarInfo{
					Name:   name.Name,
					SrcPkg: pkgName,
				}

				// 收集文档注释
				if genDecl.Doc != nil {
					info.Doc = genDecl.Doc.Text()
				} else if valueSpec.Doc != nil {
					info.Doc = valueSpec.Doc.Text()
				}

				vars = append(vars, info)
			}
		}
	}

	return vars
}

func formatTypeParams(params *ast.FieldList) string {
	var parts []string
	for _, field := range params.List {
		var names []string
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
		typeStr := exprToString(field.Type)
		if len(names) > 0 {
			parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
		} else {
			parts = append(parts, typeStr)
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatTypeArgs(params *ast.FieldList) string {
	var names []string
	for _, field := range params.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return "[" + strings.Join(names, ", ") + "]"
}

func collectTypeParamNames(params *ast.FieldList) []string {
	var names []string
	for _, field := range params.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

func formatParams(params *ast.FieldList) string {
	if params == nil {
		return ""
	}

	var parts []string
	for _, field := range params.List {
		typeStr := exprToString(field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			var names []string
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
			parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
		}
	}
	return strings.Join(parts, ", ")
}

func formatCallArgs(params *ast.FieldList) string {
	if params == nil {
		return ""
	}

	var args []string
	for _, field := range params.List {
		// 检查是否是可变参数
		_, isVariadic := field.Type.(*ast.Ellipsis)

		for _, name := range field.Names {
			if isVariadic {
				args = append(args, name.Name+"...")
			} else {
				args = append(args, name.Name)
			}
		}
	}
	return strings.Join(args, ", ")
}

func formatResults(results *ast.FieldList, srcPkg string) string {
	return formatResultsWithExcludes(results, srcPkg, nil)
}

func formatResultsWithExcludes(results *ast.FieldList, srcPkg string, excludeNames []string) string {
	if results == nil || len(results.List) == 0 {
		return ""
	}

	var parts []string
	for _, field := range results.List {
		typeStr := exprToStringWithPkgAndExcludes(field.Type, srcPkg, excludeNames)
		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			var names []string
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
			parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
		}
	}

	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func exprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	writeExpr(&buf, expr, "", nil)
	return buf.String()
}

func exprToStringWithPkg(expr ast.Expr, srcPkg string) string {
	var buf bytes.Buffer
	writeExpr(&buf, expr, srcPkg, nil)
	return buf.String()
}

func exprToStringWithPkgAndExcludes(expr ast.Expr, srcPkg string, excludeNames []string) string {
	var buf bytes.Buffer
	writeExpr(&buf, expr, srcPkg, excludeNames)
	return buf.String()
}

func isInList(name string, list []string) bool {
	for _, n := range list {
		if n == name {
			return true
		}
	}
	return false
}

func writeExpr(buf *bytes.Buffer, expr ast.Expr, srcPkg string, excludeNames []string) {
	switch e := expr.(type) {
	case *ast.Ident:
		// 如果是源包中定义的类型，添加包前缀
		// 但如果是泛型参数名称（在 excludeNames 中），则不添加前缀
		if srcPkg != "" && ast.IsExported(e.Name) && !isInList(e.Name, excludeNames) {
			buf.WriteString(srcPkg + ".")
		}
		buf.WriteString(e.Name)
	case *ast.SelectorExpr:
		writeExpr(buf, e.X, "", excludeNames)
		buf.WriteString(".")
		buf.WriteString(e.Sel.Name)
	case *ast.StarExpr:
		buf.WriteString("*")
		writeExpr(buf, e.X, srcPkg, excludeNames)
	case *ast.ArrayType:
		buf.WriteString("[]")
		writeExpr(buf, e.Elt, srcPkg, excludeNames)
	case *ast.MapType:
		buf.WriteString("map[")
		writeExpr(buf, e.Key, srcPkg, excludeNames)
		buf.WriteString("]")
		writeExpr(buf, e.Value, srcPkg, excludeNames)
	case *ast.InterfaceType:
		if e.Methods == nil || len(e.Methods.List) == 0 {
			buf.WriteString("any")
		} else {
			buf.WriteString("interface{ ")
			for i, method := range e.Methods.List {
				if i > 0 {
					buf.WriteString("; ")
				}
				// 方法名
				if len(method.Names) > 0 {
					buf.WriteString(method.Names[0].Name)
				}
				// 方法签名
				if funcType, ok := method.Type.(*ast.FuncType); ok {
					buf.WriteString("(")
					if funcType.Params != nil {
						for j, param := range funcType.Params.List {
							if j > 0 {
								buf.WriteString(", ")
							}
							writeExpr(buf, param.Type, srcPkg, excludeNames)
						}
					}
					buf.WriteString(")")
					if funcType.Results != nil && len(funcType.Results.List) > 0 {
						buf.WriteString(" ")
						if len(funcType.Results.List) == 1 && len(funcType.Results.List[0].Names) == 0 {
							writeExpr(buf, funcType.Results.List[0].Type, srcPkg, excludeNames)
						} else {
							buf.WriteString("(")
							for j, result := range funcType.Results.List {
								if j > 0 {
									buf.WriteString(", ")
								}
								writeExpr(buf, result.Type, srcPkg, excludeNames)
							}
							buf.WriteString(")")
						}
					}
				} else {
					// 嵌入类型
					writeExpr(buf, method.Type, srcPkg, excludeNames)
				}
			}
			buf.WriteString(" }")
		}
	case *ast.IndexExpr:
		writeExpr(buf, e.X, srcPkg, excludeNames)
		buf.WriteString("[")
		writeExpr(buf, e.Index, srcPkg, excludeNames)
		buf.WriteString("]")
	case *ast.IndexListExpr:
		writeExpr(buf, e.X, srcPkg, excludeNames)
		buf.WriteString("[")
		for i, index := range e.Indices {
			if i > 0 {
				buf.WriteString(", ")
			}
			writeExpr(buf, index, srcPkg, excludeNames)
		}
		buf.WriteString("]")
	case *ast.Ellipsis:
		buf.WriteString("...")
		writeExpr(buf, e.Elt, srcPkg, excludeNames)
	case *ast.FuncType:
		buf.WriteString("func(")
		if e.Params != nil {
			for i, field := range e.Params.List {
				if i > 0 {
					buf.WriteString(", ")
				}
				writeExpr(buf, field.Type, srcPkg, excludeNames)
			}
		}
		buf.WriteString(")")
		if e.Results != nil && len(e.Results.List) > 0 {
			buf.WriteString(" ")
			if len(e.Results.List) == 1 && len(e.Results.List[0].Names) == 0 {
				writeExpr(buf, e.Results.List[0].Type, srcPkg, excludeNames)
			} else {
				buf.WriteString("(")
				for i, field := range e.Results.List {
					if i > 0 {
						buf.WriteString(", ")
					}
					writeExpr(buf, field.Type, srcPkg, excludeNames)
				}
				buf.WriteString(")")
			}
		}
	case *ast.BinaryExpr:
		// 处理联合类型约束，如 ~int | ~int8 | ~int16
		writeExpr(buf, e.X, srcPkg, excludeNames)
		buf.WriteString(" " + e.Op.String() + " ")
		writeExpr(buf, e.Y, srcPkg, excludeNames)
	case *ast.UnaryExpr:
		// 处理类型约束前缀，如 ~int
		buf.WriteString(e.Op.String())
		writeExpr(buf, e.X, srcPkg, excludeNames)
	default:
		buf.WriteString(fmt.Sprintf("%T", expr))
	}
}

func isExcluded(name string, excludeList []string) bool {
	for _, ex := range excludeList {
		if name == ex {
			return true
		}
		// 支持前缀匹配（以 * 结尾）
		if strings.HasSuffix(ex, "*") && strings.HasPrefix(name, strings.TrimSuffix(ex, "*")) {
			return true
		}
	}
	return false
}

func getImportPath(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	// 查找 go.mod 文件
	modDir := absDir
	for {
		modPath := filepath.Join(modDir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			// 读取 go.mod 获取模块名
			content, err := os.ReadFile(modPath)
			if err != nil {
				return "", err
			}
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					moduleName := strings.TrimPrefix(line, "module ")
					moduleName = strings.TrimSpace(moduleName)
					relPath, err := filepath.Rel(modDir, absDir)
					if err != nil {
						return "", err
					}
					if relPath == "." {
						return moduleName, nil
					}
					return moduleName + "/" + filepath.ToSlash(relPath), nil
				}
			}
		}

		parent := filepath.Dir(modDir)
		if parent == modDir {
			break
		}
		modDir = parent
	}

	return "", fmt.Errorf("could not find go.mod file")
}

func generateCode(cfg Config, srcPkgName, srcImportPath string, types []TypeInfo, funcs []FuncInfo, vars []VarInfo) string {
	var buf bytes.Buffer

	// 头部注释
	buf.WriteString("// Code generated by export-types. DO NOT EDIT.\n")
	buf.WriteString("// Source: " + srcImportPath + "\n")
	if cfg.GenCmd != "" {
		buf.WriteString("//\n")
		buf.WriteString("//go:generate " + cfg.GenCmd + "\n")
	}
	buf.WriteString("\n")

	// 包声明
	buf.WriteString("package " + cfg.PkgName + "\n\n")

	// 导入
	buf.WriteString("import (\n")
	buf.WriteString("\t\"" + srcImportPath + "\"\n")

	// 检查所需要的包
	needFieldPkg := false
	needFieldsPkg := false
	needClausePkg := false

	// 检查函数参数和返回值
	for _, f := range funcs {
		all := f.Params + f.Results + f.TypeParams
		if strings.Contains(all, "field.") {
			needFieldPkg = true
		}
		if strings.Contains(all, "fields.") {
			needFieldsPkg = true
		}
		if strings.Contains(all, "clause.") {
			needClausePkg = true
		}
	}

	// 检查类型定义
	for _, t := range types {
		if strings.Contains(t.TypeParams, "field.") {
			needFieldPkg = true
		}
		if strings.Contains(t.TypeParams, "fields.") {
			needFieldsPkg = true
		}
		if strings.Contains(t.TypeParams, "clause.") {
			needClausePkg = true
		}
	}

	if needClausePkg {
		buf.WriteString("\t\"github.com/donutnomad/gsql/clause\"\n")
	}
	if needFieldPkg {
		buf.WriteString("\t\"github.com/donutnomad/gsql/field\"\n")
	}
	if needFieldsPkg {
		buf.WriteString("\t\"github.com/donutnomad/gsql/internal/fields\"\n")
	}
	buf.WriteString(")\n\n")

	// 构造函数（放在前面，更常用）
	if len(funcs) > 0 {
		buf.WriteString("// ==================== Constructors ====================\n\n")
		for _, f := range funcs {
			// 输出文档注释
			if f.Doc != "" {
				for _, line := range strings.Split(strings.TrimSuffix(f.Doc, "\n"), "\n") {
					buf.WriteString("// " + line + "\n")
				}
			}
			if f.IsGeneric {
				buf.WriteString(fmt.Sprintf("func %s%s(%s) %s {\n",
					f.Name, f.TypeParams, f.Params, f.Results))
				buf.WriteString(fmt.Sprintf("\treturn %s.%s%s(%s)\n",
					srcPkgName, f.Name, f.TypeArgs, f.CallArgs))
				buf.WriteString("}\n\n")
			} else {
				buf.WriteString(fmt.Sprintf("func %s(%s) %s {\n",
					f.Name, f.Params, f.Results))
				buf.WriteString(fmt.Sprintf("\treturn %s.%s(%s)\n",
					srcPkgName, f.Name, f.CallArgs))
				buf.WriteString("}\n\n")
			}
		}
	}

	// 类型别名（放在后面，使用紧凑格式）
	if len(types) > 0 {
		buf.WriteString("// ==================== Type Aliases ====================\n\n")
		buf.WriteString("type (\n")
		for _, t := range types {
			if t.IsGeneric {
				buf.WriteString(fmt.Sprintf("\t%s%s = %s.%s%s\n",
					t.Name, t.TypeParams, srcPkgName, t.Name, t.TypeArgs))
			} else {
				buf.WriteString(fmt.Sprintf("\t%s = %s.%s\n",
					t.Name, srcPkgName, t.Name))
			}
		}
		buf.WriteString(")\n")
	}

	// 变量别名
	if len(vars) > 0 {
		buf.WriteString("\n// ==================== Variables ====================\n\n")
		buf.WriteString("var (\n")
		for _, v := range vars {
			// 输出文档注释
			if v.Doc != "" {
				for _, line := range strings.Split(strings.TrimSuffix(v.Doc, "\n"), "\n") {
					buf.WriteString("\t// " + line + "\n")
				}
			}
			buf.WriteString(fmt.Sprintf("\t%s = %s.%s\n", v.Name, srcPkgName, v.Name))
		}
		buf.WriteString(")\n")
	}

	return buf.String()
}

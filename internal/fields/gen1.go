//go:build ignore

//go:generate go run gen1.go

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// ==================== 数据结构 ====================

// GenDirective 代码生成指令，从 @gen 注释解析
type GenDirective struct {
	Public      string   // 公开方法名
	Return      string   // 返回类型 ("self" 表示返回接收者类型)
	Constructor string   // 构造函数 (可选，自动推导)
	For         []string // 适用的表达式类型 (空表示所有嵌入此类型的表达式)
	Exclude     []string // 排除的表达式类型
	Direct      bool     // 直接返回，不用构造函数包装
	Void        bool     // 无返回值
	Param       string   // 覆盖参数类型 (如 "int64", "IntExpr[T]")
	ParamName   string   // 参数名 (默认使用原始参数名)
}

// MethodInfo 从源码解析的方法信息
type MethodInfo struct {
	ReceiverType string         // 接收者类型名 (如 arithmeticSql)
	InnerName    string         // 内部方法名 (如 addExpr)
	Params       string         // 参数列表 (如 "value any")
	Args         string         // 调用参数 (如 "value")
	Comments     []string       // 注释行 (不含 @gen)
	Directives   []GenDirective // 生成指令
}

// TypeInfo 表达式类型信息
type TypeInfo struct {
	Name         string // 类型名 (如 IntExpr)
	TypeParam    string // 泛型参数 (如 [T])
	Constructor  string // 构造函数名 (如 NewIntExpr)
	FileName     string // 源文件名 (不含 .go)
	Embeddings   []string // 嵌入的类型列表
	DefaultParam string // 默认泛型参数 (如 [int], [string])
}

// GeneratedMethod 要生成的方法
type GeneratedMethod struct {
	TypeName        string
	TypeParam       string
	MethodName      string
	InnerName       string
	Params          string   // 生成的参数列表 (如 "value int64")
	Args            string   // 调用内部方法的参数 (如 "value")
	ReturnType      string
	Constructor     string
	Comments        []string
	Direct          bool
	Void            bool
	VariadicConvert bool   // 是否需要变参转换（从具体类型到 any）
	VariadicArgName string // 变参参数名
}

// ==================== 配置 ====================

// 源文件列表 (包含内部方法定义，带 @gen 注解的方法)
var sourceFiles = []string{
	"numeric_base.go",
}

// ==================== 主程序 ====================

func main() {
	// 1. 扫描目录，解析所有带 @gentype 注解的表达式类型
	types := scanAndParseTypes()
	if len(types) == 0 {
		fmt.Fprintln(os.Stderr, "No types found (ensure structs have @gentype annotation)")
		os.Exit(1)
	}

	// 2. 解析所有源文件获取方法信息
	methods := parseSourceFiles()
	if len(methods) == 0 {
		fmt.Fprintln(os.Stderr, "No methods found")
		os.Exit(1)
	}

	// 3. 构建默认泛型参数映射
	defaultTypeParams := buildDefaultTypeParams(types)

	// 4. 为每个类型生成方法
	generated := generateMethods(types, methods, defaultTypeParams)

	// 5. 按文件分组并写入
	writeGeneratedFiles(types, generated)

	// 6. 清理旧文件
	os.Remove("generate1.go")
}

// ==================== 类型扫描与解析 ====================

// scanAndParseTypes 扫描当前目录的所有 .go 文件，查找带 @gentype 注解的结构体
func scanAndParseTypes() []TypeInfo {
	var types []TypeInfo

	// 获取当前目录下所有 .go 文件
	files, err := filepath.Glob("*.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing files: %v\n", err)
		return nil
	}

	for _, fileName := range files {
		// 跳过生成的文件和测试文件
		if strings.HasSuffix(fileName, "_generate.go") ||
			strings.HasSuffix(fileName, "_test.go") ||
			fileName == "gen1.go" ||
			fileName == "gen.go" {
			continue
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", fileName, err)
			continue
		}

		// 遍历所有声明
		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// 检查是否有 @gentype 注解
				gentypeDirective := parseGentypeDirective(genDecl.Doc)
				if gentypeDirective == nil {
					continue
				}

				info := TypeInfo{
					Name:         typeSpec.Name.Name,
					FileName:     strings.TrimSuffix(fileName, ".go"),
					DefaultParam: gentypeDirective.DefaultParam,
				}

				// 解析泛型参数
				if typeSpec.TypeParams != nil && len(typeSpec.TypeParams.List) > 0 {
					info.TypeParam = "[T]"
				}

				// 推导构造函数名
				info.Constructor = "New" + info.Name

				// 解析嵌入字段
				for _, field := range structType.Fields.List {
					if len(field.Names) > 0 {
						continue // 非嵌入字段
					}
					embName := getTypeName(field.Type)
					if embName != "" {
						info.Embeddings = append(info.Embeddings, embName)
					}
				}

				types = append(types, info)
			}
		}
	}

	return types
}

// GentypeDirective @gentype 注解的解析结果
type GentypeDirective struct {
	DefaultParam string // 默认泛型参数，如 [int], [string]
}

// parseGentypeDirective 解析 @gentype 注解
// 格式: @gentype default=[int]
func parseGentypeDirective(doc *ast.CommentGroup) *GentypeDirective {
	if doc == nil {
		return nil
	}

	for _, comment := range doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimSpace(text)

		if !strings.HasPrefix(text, "@gentype") {
			continue
		}

		d := &GentypeDirective{}

		// 解析参数
		text = strings.TrimPrefix(text, "@gentype")
		text = strings.TrimSpace(text)

		// 使用正则解析 default=[xxx]
		re := regexp.MustCompile(`default=(\[[^\]]+\])`)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			d.DefaultParam = matches[1]
		}

		return d
	}

	return nil
}

// buildDefaultTypeParams 从类型列表构建默认泛型参数映射
func buildDefaultTypeParams(types []TypeInfo) map[string]string {
	result := make(map[string]string)
	for _, t := range types {
		if t.DefaultParam != "" {
			result[t.Name] = t.DefaultParam
		}
	}
	return result
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

// ==================== 方法解析 ====================

func parseSourceFiles() []MethodInfo {
	var methods []MethodInfo

	for _, fileName := range sourceFiles {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", fileName, err)
			continue
		}

		for _, decl := range node.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}

			// 获取接收者类型
			recvType := getTypeName(fn.Recv.List[0].Type)
			if recvType == "" {
				continue
			}

			// 解析注释
			var comments []string
			var directives []GenDirective
			if fn.Doc != nil {
				for _, c := range fn.Doc.List {
					text := strings.TrimPrefix(c.Text, "//")
					text = strings.TrimSpace(text)
					if strings.HasPrefix(text, "@gen") {
						if d := parseGenDirective(text); d != nil {
							directives = append(directives, *d)
						}
					} else if text != "" {
						comments = append(comments, text)
					}
				}
			}

			// 没有 @gen 指令的方法跳过（完全依赖注解驱动）
			if len(directives) == 0 {
				continue
			}

			// 解析参数
			params, args := parseParams(fn.Type.Params)

			methods = append(methods, MethodInfo{
				ReceiverType: recvType,
				InnerName:    fn.Name.Name,
				Params:       params,
				Args:         args,
				Comments:     comments,
				Directives:   directives,
			})
		}
	}

	return methods
}

// parseGenDirective 解析 @gen 指令
// 格式: @gen public=Name return=Type for=Type1,Type2 exclude=Type3 param=int64 paramName=value
func parseGenDirective(text string) *GenDirective {
	text = strings.TrimPrefix(text, "@gen")
	text = strings.TrimSpace(text)

	d := &GenDirective{}

	// 使用正则解析键值对
	re := regexp.MustCompile(`(\w+)=([^\s]+)`)
	matches := re.FindAllStringSubmatch(text, -1)

	for _, m := range matches {
		key, value := m[1], m[2]
		switch key {
		case "public":
			d.Public = value
		case "return":
			d.Return = value
		case "constructor":
			d.Constructor = value
		case "for":
			// 去掉方括号后按逗号分隔，如 [IntExpr,FloatExpr] -> ["IntExpr", "FloatExpr"]
			d.For = strings.Split(strings.Trim(value, "[]"), ",")
		case "exclude":
			d.Exclude = strings.Split(strings.Trim(value, "[]"), ",")
		case "direct":
			d.Direct = value == "true"
		case "void":
			d.Void = value == "true"
		case "param":
			d.Param = value
		case "paramName":
			d.ParamName = value
		}
	}

	if d.Public == "" {
		return nil
	}

	return d
}

func parseParams(params *ast.FieldList) (paramStr, argStr string) {
	if params == nil || len(params.List) == 0 {
		return "", ""
	}

	var paramParts []string
	var argParts []string

	for _, field := range params.List {
		typeStr := exprToString(field.Type)
		for _, name := range field.Names {
			paramParts = append(paramParts, name.Name+" "+typeStr)
			// 处理可变参数
			if strings.HasPrefix(typeStr, "...") {
				argParts = append(argParts, name.Name+"...")
			} else {
				argParts = append(argParts, name.Name)
			}
		}
	}

	return strings.Join(paramParts, ", "), strings.Join(argParts, ", ")
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		return "[" + exprToString(t.Len) + "]" + exprToString(t.Elt)
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	case *ast.IndexExpr:
		return exprToString(t.X) + "[" + exprToString(t.Index) + "]"
	case *ast.BasicLit:
		return t.Value
	}
	return "any"
}

// ==================== 代码生成 ====================

func generateMethods(types []TypeInfo, methods []MethodInfo, defaultTypeParams map[string]string) map[string][]GeneratedMethod {
	result := make(map[string][]GeneratedMethod)

	// 构建类型名到类型信息的映射
	typeMap := make(map[string]TypeInfo)
	for _, t := range types {
		typeMap[t.Name] = t
	}

	// 构建嵌入类型到表达式类型的映射
	embeddingToTypes := make(map[string][]string)
	for _, t := range types {
		for _, emb := range t.Embeddings {
			embeddingToTypes[emb] = append(embeddingToTypes[emb], t.Name)
		}
	}

	for _, method := range methods {
		for _, directive := range method.Directives {
			// 确定适用的类型
			var targetTypes []string
			if len(directive.For) > 0 {
				targetTypes = directive.For
			} else {
				targetTypes = embeddingToTypes[method.ReceiverType]
			}

			// 排除指定类型
			if len(directive.Exclude) > 0 {
				excludeMap := make(map[string]bool)
				for _, e := range directive.Exclude {
					excludeMap[e] = true
				}
				var filtered []string
				for _, t := range targetTypes {
					if !excludeMap[t] {
						filtered = append(filtered, t)
					}
				}
				targetTypes = filtered
			}

			// 为每个目标类型生成方法
			for _, typeName := range targetTypes {
				typeInfo, ok := typeMap[typeName]
				if !ok {
					continue
				}

				gm := GeneratedMethod{
					TypeName:   typeInfo.Name,
					TypeParam:  typeInfo.TypeParam,
					MethodName: directive.Public,
					InnerName:  method.InnerName,
					Params:     method.Params,
					Args:       method.Args,
					Comments:   method.Comments,
					Direct:     directive.Direct,
					Void:       directive.Void,
				}

				// 处理参数类型覆盖
				if directive.Param != "" {
					paramName := directive.ParamName
					if paramName == "" {
						// 从原始参数中提取第一个参数名
						paramName = extractFirstParamName(method.Args)
					}
					// 处理泛型参数替换 [T] -> 调用者的泛型参数
					paramType := normalizeReturnType(directive.Param, typeInfo.TypeParam, defaultTypeParams)
					// 检查原始参数是否是变参（args 以 ... 结尾）
					isVariadic := strings.HasSuffix(method.Args, "...")
					if isVariadic {
						gm.Params = paramName + " ..." + paramType
						gm.Args = "_anyValues..."
						gm.VariadicConvert = true
						gm.VariadicArgName = paramName
					} else {
						gm.Params = paramName + " " + paramType
						gm.Args = paramName
					}
				}

				// 确定返回类型
				if directive.Void {
					// 无返回值
				} else if directive.Direct {
					gm.ReturnType = directive.Return
				} else if directive.Return == "self" || directive.Return == "" {
					// 返回自身类型
					gm.ReturnType = typeInfo.Name + typeInfo.TypeParam
					gm.Constructor = typeInfo.Constructor + typeInfo.TypeParam
				} else {
					// 指定返回类型
					gm.ReturnType = normalizeReturnType(directive.Return, typeInfo.TypeParam, defaultTypeParams)
					gm.Constructor = directive.Constructor
					if gm.Constructor == "" {
						gm.Constructor = deriveConstructor(gm.ReturnType)
					}
				}

				result[typeInfo.FileName] = append(result[typeInfo.FileName], gm)
			}
		}
	}

	return result
}

// extractFirstParamName 从参数列表字符串中提取第一个参数名
// 如 "value, other..." -> "value"
func extractFirstParamName(args string) string {
	if args == "" {
		return "value"
	}
	// 去除可能的 ... 后缀和空格
	args = strings.TrimSuffix(args, "...")
	args = strings.TrimSpace(args)
	// 按逗号分割取第一个
	if idx := strings.Index(args, ","); idx > 0 {
		return strings.TrimSpace(args[:idx])
	}
	return args
}

func normalizeReturnType(ret string, callerTypeParam string, defaultTypeParams map[string]string) string {
	// 如果返回类型包含 [T]，用调用者的泛型参数替换
	if strings.Contains(ret, "[T]") {
		return strings.Replace(ret, "[T]", callerTypeParam, 1)
	}
	// 如果已经有其他泛型参数，直接返回
	if strings.Contains(ret, "[") {
		return ret
	}
	// 添加默认泛型参数
	if param, ok := defaultTypeParams[ret]; ok {
		return ret + param
	}
	return ret
}

func deriveConstructor(returnType string) string {
	// 从返回类型推导构造函数
	// IntExpr[int8] -> NewIntExpr[int8]
	if idx := strings.Index(returnType, "["); idx > 0 {
		baseName := returnType[:idx]
		typeParam := returnType[idx:]
		return "New" + baseName + typeParam
	}
	return "New" + returnType
}

// ==================== 文件写入 ====================

const methodTemplate = `{{range .Comments}}
// {{.}}{{end}}
func (e {{.TypeName}}{{.TypeParam}}) {{.MethodName}}({{.Params}}){{if .ReturnType}} {{.ReturnType}}{{end}} {
{{if .VariadicConvert}}	_anyValues := make([]any, len({{.VariadicArgName}}))
	for _i, _v := range {{.VariadicArgName}} {
		_anyValues[_i] = _v
	}
{{end}}{{if .Void}}	e.{{.InnerName}}({{.Args}})
{{else if .Direct}}	return e.{{.InnerName}}({{.Args}})
{{else}}	return {{.Constructor}}(e.{{.InnerName}}({{.Args}}))
{{end}}}
`

func writeGeneratedFiles(types []TypeInfo, generated map[string][]GeneratedMethod) {
	tmpl := template.Must(template.New("method").Parse(methodTemplate))

	// 按文件名分组类型
	fileToTypes := make(map[string][]TypeInfo)
	for _, t := range types {
		fileToTypes[t.FileName] = append(fileToTypes[t.FileName], t)
	}

	for fileName, methods := range generated {
		if len(methods) == 0 {
			continue
		}

		var buf bytes.Buffer

		// 检查需要的导入
		needClauseImport := false
		needFieldImport := false
		for _, m := range methods {
			if strings.Contains(m.Params, "clause.") || strings.Contains(m.ReturnType, "clause.") {
				needClauseImport = true
			}
			if strings.Contains(m.Params, "field.") || strings.Contains(m.ReturnType, "field.") {
				needFieldImport = true
			}
		}

		// 写入文件头
		buf.WriteString("// Code generated by gen1.go; DO NOT EDIT.\n\npackage fields\n")
		if needClauseImport || needFieldImport {
			buf.WriteString("\nimport (\n")
			if needClauseImport {
				buf.WriteString("\t\"github.com/donutnomad/gsql/clause\"\n")
			}
			if needFieldImport {
				buf.WriteString("\t\"github.com/donutnomad/gsql/field\"\n")
			}
			buf.WriteString(")\n")
		}
		buf.WriteString("\n")

		// 按类型分组写入
		typesMethods := make(map[string][]GeneratedMethod)
		for _, m := range methods {
			typesMethods[m.TypeName] = append(typesMethods[m.TypeName], m)
		}

		for _, typeInfo := range fileToTypes[fileName] {
			typeMethods := typesMethods[typeInfo.Name]
			if len(typeMethods) == 0 {
				continue
			}

			fmt.Fprintf(&buf, "// ==================== %s 生成的方法 ====================\n", typeInfo.Name)

			for _, m := range typeMethods {
				if err := tmpl.Execute(&buf, m); err != nil {
					fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
					os.Exit(1)
				}
			}
			buf.WriteString("\n")
		}

		// 写入文件
		outputFile := fileName + "_generate.go"
		if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outputFile, err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s successfully\n", outputFile)
	}
}

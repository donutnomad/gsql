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
	"strings"
	"text/template"
)

// MethodDef 方法定义
type MethodDef struct {
	Name        string   // 公开方法名 (首字母大写)，InnerName 自动生成为 首字母小写+Expr
	InnerName   string   // 内部方法名 (可选，为空时自动生成)
	Comments    []string // 注释行
	Params      string   // 参数列表 (如 "defaultValue any")
	Args        string   // 调用参数 (如 "defaultValue")
	ReturnType  string   // 返回类型 (空表示返回自身类型)
	Constructor string   // 返回类型的构造函数 (空表示使用接收者的构造函数)
	Direct      bool     // true 表示直接调用内部方法，不用构造函数包装
	Void        bool     // true 表示无返回值
}

// getInnerName 获取内部方法名
// 如果 InnerName 已设置则直接返回，否则自动生成：首字母小写 + "Expr"
func (m MethodDef) getInnerName() string {
	if m.InnerName != "" {
		return m.InnerName
	}
	// 自动生成: Name 首字母小写 + "Expr"
	if len(m.Name) == 0 {
		return ""
	}
	return strings.ToLower(m.Name[:1]) + m.Name[1:] + "Expr"
}

// TypeDef 类型定义
type TypeDef struct {
	Name        string // 类型名 (如 IntExpr)
	TypeParam   string // 泛型参数 (如 [T])
	Constructor string // 构造函数名
	FileName    string // 对应的源文件名 (不含 .go 后缀)
}

// EmbeddedMethods 嵌入类型及其方法映射
// InnerName 可省略，默认自动生成为: 首字母小写(Name) + "Expr"
var embeddedMethods = map[string][]MethodDef{
	// 基础表达式方法 (Build, ToExpr, As)
	// Build 方法用于解决多个嵌入 clause.Expression 导致的歧义问题
	"baseExprSql": {
		{Name: "Build", Params: "builder clause.Builder", Args: "builder", Void: true},
		{Name: "ToExpr", ReturnType: "clause.Expression", Direct: true},
		{Name: "As", Params: "alias string", Args: "alias", ReturnType: "field.IField", Direct: true},
	},
	"arithmeticSql": {
		{Name: "Add", Params: "value any", Args: "value"},
		{Name: "Sub", Params: "value any", Args: "value"},
		{Name: "Mul", Params: "value any", Args: "value"},
		{Name: "Div", Params: "value any", Args: "value"},
		{Name: "Neg"},
		{Name: "Mod", Params: "value any", Args: "value"},
	},
	"mathFuncSql": {
		{Name: "Abs"},
		{Name: "Round", Params: "decimals ...int", Args: "decimals..."},
		{Name: "Truncate", Params: "decimals int", Args: "decimals"},
	},
	// IntExpr 特有的数学函数 (返回自身类型)
	"mathFuncIntSql": {
		{Name: "Sign", ReturnType: "IntExpr[int8]", Constructor: "NewIntExpr[int8]"},
		{Name: "Ceil"},
		{Name: "Floor"},
		{Name: "Pow", Params: "exponent int", Args: "float64(exponent)", ReturnType: "FloatExpr[float64]", Constructor: "NewFloatExpr[float64]"},
		{Name: "Sqrt", ReturnType: "FloatExpr[float64]", Constructor: "NewFloatExpr[float64]"},
		{Name: "Log", ReturnType: "FloatExpr[float64]", Constructor: "NewFloatExpr[float64]"},
		{Name: "Log10", ReturnType: "FloatExpr[float64]", Constructor: "NewFloatExpr[float64]"},
		{Name: "Log2", ReturnType: "FloatExpr[float64]", Constructor: "NewFloatExpr[float64]"},
		{Name: "Exp", ReturnType: "FloatExpr[float64]", Constructor: "NewFloatExpr[float64]"},
	},
	// FloatExpr/DecimalExpr 的数学函数 (Sign/Ceil/Floor 返回不同类型)
	"mathFuncFloatSql": {
		{Name: "Sign", ReturnType: "IntExpr[int8]", Constructor: "NewIntExpr[int8]"},
		{Name: "Ceil", ReturnType: "IntExpr[int64]", Constructor: "NewIntExpr[int64]"},
		{Name: "Floor", ReturnType: "IntExpr[int64]", Constructor: "NewIntExpr[int64]"},
		{Name: "Pow", Params: "exponent float64", Args: "exponent"},
		{Name: "Sqrt"},
		{Name: "Log"},
		{Name: "Log10"},
		{Name: "Log2"},
		{Name: "Exp"},
	},
	"nullCondFuncSql": {
		{Name: "IfNull", Params: "defaultValue any", Args: "defaultValue"},
		{Name: "Coalesce", Params: "values ...any", Args: "values..."},
		{Name: "Nullif", Params: "value any", Args: "value"},
	},
	"numericCondFuncSql": {
		{Name: "Greatest", Params: "values ...any", Args: "values..."},
		{Name: "Least", Params: "values ...any", Args: "values..."},
	},
	"bitOpSql": {
		{Name: "BitAnd", Params: "value any", Args: "value"},
		{Name: "BitOr", Params: "value any", Args: "value"},
		{Name: "BitXor", Params: "value any", Args: "value"},
		{Name: "BitNot"},
		{Name: "LeftShift", Params: "n int", Args: "n"},
		{Name: "RightShift", Params: "n int", Args: "n"},
		{Name: "IntDiv", Params: "value any", Args: "value"},
	},
	// IntExpr 的聚合函数 (Avg 返回 FloatExpr)
	"aggregateIntSql": {
		{Name: "Sum"},
		{Name: "Avg", ReturnType: "FloatExpr[float64]", Constructor: "NewFloatExpr[float64]"},
		{Name: "Max"},
		{Name: "Min"},
	},
	// 三角函数 (FloatExpr 专用)
	"trigFuncSql": {
		{Name: "Sin"},
		{Name: "Cos"},
		{Name: "Tan"},
		{Name: "Asin"},
		{Name: "Acos"},
		{Name: "Atan"},
		{Name: "Radians"},
		{Name: "Degrees"},
	},
	// FloatExpr/DecimalExpr 的聚合函数 (返回自身类型)
	"aggregateFloatSql": {
		{Name: "Sum"},
		{Name: "Avg"},
		{Name: "Max"},
		{Name: "Min"},
	},
	// 日期时间类型的聚合函数 (只有 Max/Min)
	"aggregateDateSql": {
		{Name: "Max"},
		{Name: "Min"},
	},
	// 日期提取函数 (DateExpr, DateTimeExpr, TimestampExpr 共用)
	"dateExtractSql": {
		{Name: "Year", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "Month", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "Day", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "DayOfMonth", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "DayOfWeek", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "DayOfYear", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "Week", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "WeekOfYear", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "Quarter", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
	},
	// 时间提取函数 (DateTimeExpr, TimeExpr, TimestampExpr 共用)
	"timeExtractSql": {
		{Name: "Hour", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "Minute", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "Second", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
		{Name: "Microsecond", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
	},
	// 日期时间间隔运算 (DateExpr, DateTimeExpr, TimeExpr, TimestampExpr 共用)
	"dateIntervalSql": {
		{Name: "AddInterval", Params: "interval string", Args: "interval"},
		{Name: "SubInterval", Params: "interval string", Args: "interval"},
	},
	// 日期差值计算 (DateExpr, DateTimeExpr, TimestampExpr)
	"dateDiffSql": {
		{Name: "DateDiff", Params: "other clause.Expression", Args: "other", ReturnType: "IntExpr[int]", Constructor: "NewIntExpr[int]"},
	},
	// 时间差值计算 (DateTimeExpr, TimeExpr, TimestampExpr)
	// 注意: TimeExpr 的 TimeDiff 返回自身类型，需要特殊处理
	"timeDiffSql": {
		{Name: "TimeDiff", Params: "other clause.Expression", Args: "other", ReturnType: "TimeExpr[string]", Constructor: "NewTimeExpr[string]"},
	},
	// TimeExpr 专用的 TimeDiff (返回自身类型)
	"timeDiffSelfSql": {
		{Name: "TimeDiff", Params: "other clause.Expression", Args: "other"},
	},
	// 时间戳差值计算 (DateTimeExpr, TimestampExpr)
	"timestampDiffSql": {
		{Name: "TimestampDiff", Params: "unit string, other clause.Expression", Args: "unit, other", ReturnType: "IntExpr[int64]", Constructor: "NewIntExpr[int64]"},
	},
	// 日期格式化 (DateExpr, DateTimeExpr, TimestampExpr)
	// 注意: Format 的内部方法名是 dateFormatExpr，不是默认的 formatExpr
	"dateFormatSql": {
		{Name: "Format", InnerName: "dateFormatExpr", Params: "format string", Args: "format", ReturnType: "TextExpr[string]", Constructor: "NewTextExpr[string]"},
	},
	// 时间格式化 (TimeExpr)
	// 注意: Format 的内部方法名是 timeFormatExpr
	"timeFormatSql": {
		{Name: "Format", InnerName: "timeFormatExpr", Params: "format string", Args: "format", ReturnType: "TextExpr[string]", Constructor: "NewTextExpr[string]"},
	},
	// 日期时间转换 (DateTimeExpr, TimestampExpr)
	// 注意: Date/Time 的内部方法名是 extractDateExpr/extractTimeExpr
	"dateConversionSql": {
		{Name: "Date", InnerName: "extractDateExpr", ReturnType: "DateExpr[string]", Constructor: "NewDateExpr[string]"},
		{Name: "Time", InnerName: "extractTimeExpr", ReturnType: "TimeExpr[string]", Constructor: "NewTimeExpr[string]"},
	},
	// Unix 时间戳转换 (DateExpr, DateTimeExpr, TimestampExpr)
	"unixTimestampSql": {
		{Name: "UnixTimestamp", ReturnType: "IntExpr[int64]", Constructor: "NewIntExpr[int64]"},
	},
}

// 类型配置
var typeConfigs = []TypeDef{
	{Name: "IntExpr", TypeParam: "[T]", Constructor: "NewIntExpr", FileName: "int"},
	{Name: "FloatExpr", TypeParam: "[T]", Constructor: "NewFloatExpr", FileName: "float"},
	{Name: "DecimalExpr", TypeParam: "[T]", Constructor: "NewDecimalExpr", FileName: "decimal"},
	{Name: "TextExpr", TypeParam: "[T]", Constructor: "NewTextExpr", FileName: "text"},
	{Name: "DateExpr", TypeParam: "[T]", Constructor: "NewDateExpr", FileName: "date"},
	{Name: "DateTimeExpr", TypeParam: "[T]", Constructor: "NewDateTimeExpr", FileName: "datetime"},
	{Name: "TimeExpr", TypeParam: "[T]", Constructor: "NewTimeExpr", FileName: "time"},
	{Name: "TimestampExpr", TypeParam: "[T]", Constructor: "NewTimestampExpr", FileName: "timestamp"},
	{Name: "YearExpr", TypeParam: "[T]", Constructor: "NewYearExpr", FileName: "year"},
}

// embeddingAdditions 嵌入类型追加配置
// 用于处理"虚拟嵌入"：实际嵌入类型 -> 额外的代码生成类型列表
// 这些追加类型定义了特殊的返回值类型方法
var embeddingAdditions = map[string]map[string][]string{
	"IntExpr": {
		"mathFuncSql":  {"mathFuncIntSql"},  // 追加 IntExpr 特有的数学函数
		"aggregateSql": {"aggregateIntSql"}, // 替换为 IntExpr 特有的聚合函数
	},
	"FloatExpr": {
		"mathFuncSql":  {"mathFuncFloatSql"},  // 追加 Float 的数学函数
		"aggregateSql": {"aggregateFloatSql"}, // 替换为 Float 的聚合函数
	},
	"DecimalExpr": {
		"mathFuncSql":  {"mathFuncFloatSql"},
		"aggregateSql": {"aggregateFloatSql"},
	},
	"TimeExpr": {
		"aggregateSql": {"aggregateDateSql"},  // 替换为只有 Max/Min 的聚合函数
		"timeDiffSql":  {"timeDiffSelfSql"},   // 替换为返回自身类型的 TimeDiff
	},
	"DateExpr": {
		"aggregateSql": {"aggregateDateSql"},
	},
	"DateTimeExpr": {
		"aggregateSql": {"aggregateDateSql"},
	},
	"TimestampExpr": {
		"aggregateSql": {"aggregateDateSql"},
	},
	"YearExpr": {
		"aggregateSql": {"aggregateDateSql"},
	},
}

// embeddingSkips 需要跳过的嵌入类型
// 当某个嵌入类型被完全替换时（如 aggregateSql 被替换为 aggregateIntSql），原嵌入需要跳过
var embeddingSkips = map[string]map[string]bool{
	"IntExpr": {
		"aggregateSql": true, // 跳过基础聚合，使用 aggregateIntSql
	},
	"FloatExpr": {
		"aggregateSql": true,
	},
	"DecimalExpr": {
		"aggregateSql": true,
	},
	"TimeExpr": {
		"aggregateSql": true,
		"timeDiffSql":  true, // 跳过基础 timeDiffSql，使用 timeDiffSelfSql
	},
	"DateExpr": {
		"aggregateSql": true,
	},
	"DateTimeExpr": {
		"aggregateSql": true,
	},
	"TimestampExpr": {
		"aggregateSql": true,
	},
	"YearExpr": {
		"aggregateSql": true,
	},
}

const methodTemplate = `{{range .Comments}}
// {{.}}{{end}}
func (e {{.TypeName}}{{.TypeParam}}) {{.MethodName}}({{.Params}}){{if .ReturnType}} {{.ReturnType}}{{end}} {
{{if .Void}}	e.{{.InnerName}}({{.Args}})
{{else if .Direct}}	return e.{{.InnerName}}({{.Args}})
{{else}}	return {{.Constructor}}(e.{{.InnerName}}({{.Args}}))
{{end}}}
`

func main() {
	// 解析源文件获取注释
	comments := parseComments("numeric_base.go")

	// 为每个方法填充注释
	for embType, methods := range embeddedMethods {
		for i := range methods {
			innerName := methods[i].getInnerName()
			key := embType + "." + innerName
			if c, ok := comments[key]; ok {
				embeddedMethods[embType][i].Comments = c
			}
			// 如果没找到，尝试用基础类型的注释
			if len(embeddedMethods[embType][i].Comments) == 0 {
				// 尝试 mathFuncSql
				key = "mathFuncSql." + innerName
				if c, ok := comments[key]; ok {
					embeddedMethods[embType][i].Comments = c
				}
			}
			if len(embeddedMethods[embType][i].Comments) == 0 {
				// 尝试 aggregateSql
				key = "aggregateSql." + innerName
				if c, ok := comments[key]; ok {
					embeddedMethods[embType][i].Comments = c
				}
			}
		}
	}

	tmpl := template.Must(template.New("method").Parse(methodTemplate))

	// 按文件名分组类型
	fileTypes := make(map[string][]TypeDef)
	for _, typeDef := range typeConfigs {
		fileTypes[typeDef.FileName] = append(fileTypes[typeDef.FileName], typeDef)
	}

	// 为每个文件生成代码
	for fileName, types := range fileTypes {
		var buf bytes.Buffer
		needClauseImport := false
		needFieldImport := false

		// 检查是否需要导入 clause 和 field 包
		for _, typeDef := range types {
			embeddings := getTypeEmbeddings(typeDef.Name, typeDef.FileName)
			if len(embeddings) == 0 {
				continue
			}
			for _, embType := range embeddings {
				methods, ok := embeddedMethods[embType]
				if !ok {
					continue
				}
				for _, method := range methods {
					if strings.Contains(method.Params, "clause.") || strings.Contains(method.ReturnType, "clause.") {
						needClauseImport = true
					}
					if strings.Contains(method.Params, "field.") || strings.Contains(method.ReturnType, "field.") {
						needFieldImport = true
					}
				}
			}
		}

		buf.WriteString(`// Code generated by gen1.go; DO NOT EDIT.

package fields
`)
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

		for _, typeDef := range types {
			embeddings := getTypeEmbeddings(typeDef.Name, typeDef.FileName)
			if len(embeddings) == 0 {
				continue
			}

			fmt.Fprintf(&buf, "// ==================== %s 生成的方法 ====================\n", typeDef.Name)

			for _, embType := range embeddings {
				methods, ok := embeddedMethods[embType]
				if !ok {
					continue
				}

				for _, method := range methods {
					// 确定返回类型
					returnType := method.ReturnType
					constructor := method.Constructor
					if returnType == "" && !method.Void {
						// 返回自身类型（Void 方法除外）
						returnType = typeDef.Name + typeDef.TypeParam
						constructor = typeDef.Constructor + typeDef.TypeParam
					}

					data := map[string]any{
						"TypeName":    typeDef.Name,
						"TypeParam":   typeDef.TypeParam,
						"MethodName":  method.Name,
						"Params":      method.Params,
						"Args":        method.Args,
						"ReturnType":  returnType,
						"Constructor": constructor,
						"InnerName":   method.getInnerName(),
						"Comments":    method.Comments,
						"Direct":      method.Direct,
						"Void":        method.Void,
					}
					if err := tmpl.Execute(&buf, data); err != nil {
						fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
						os.Exit(1)
					}
				}
			}
			buf.WriteString("\n")
		}

		// 写入文件
		outputFile := fileName + "_generate.go"
		if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", outputFile, err)
			os.Exit(1)
		}

		fmt.Printf("Generated %s successfully\n", outputFile)
	}

	// 删除旧的 generate1.go 文件
	if err := os.Remove("generate1.go"); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove generate1.go: %v\n", err)
	}
}

// parseStructEmbeddings 解析源文件中的结构体嵌入字段
// 返回类型名 -> 嵌入字段名列表
func parseStructEmbeddings(filename string) map[string][]string {
	result := make(map[string][]string)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", filename, err)
		return result
	}

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

			typeName := typeSpec.Name.Name
			var embeddings []string

			for _, field := range structType.Fields.List {
				// 嵌入字段没有名字
				if len(field.Names) > 0 {
					continue
				}

				// 获取嵌入类型名
				var embName string
				switch t := field.Type.(type) {
				case *ast.Ident:
					embName = t.Name
				case *ast.IndexExpr:
					// 泛型类型如 numericComparableImpl[T]
					if ident, ok := t.X.(*ast.Ident); ok {
						embName = ident.Name
					}
				}

				if embName != "" {
					embeddings = append(embeddings, embName)
				}
			}

			if len(embeddings) > 0 {
				result[typeName] = embeddings
			}
		}
	}

	return result
}

// getTypeEmbeddings 获取类型的嵌入字段列表，并应用配置
// 返回用于代码生成的嵌入类型列表
func getTypeEmbeddings(typeName, fileName string) []string {
	// 解析源文件获取实际嵌入
	structEmbeddings := parseStructEmbeddings(fileName + ".go")
	actualEmbeddings := structEmbeddings[typeName]

	// 获取该类型的配置
	additions := embeddingAdditions[typeName]
	skips := embeddingSkips[typeName]

	var result []string
	for _, emb := range actualEmbeddings {
		// 检查是否需要跳过
		if skips != nil && skips[emb] {
			// 跳过此嵌入，但添加其替代项
			if addList, ok := additions[emb]; ok {
				result = append(result, addList...)
			}
			continue
		}

		// 检查是否在 embeddedMethods 中定义了
		if _, ok := embeddedMethods[emb]; ok {
			result = append(result, emb)
		}

		// 检查是否有追加项
		if addList, ok := additions[emb]; ok {
			result = append(result, addList...)
		}
		// 否则跳过（如 pointerExprImpl, castSql 等不需要生成方法的类型）
	}

	return result
}

// parseComments 解析源文件中的方法注释
func parseComments(filename string) map[string][]string {
	result := make(map[string][]string)

	// 需要解析注释的类型（包括用于注释回退的基础类型）
	commentTypes := make(map[string]bool)
	for k := range embeddedMethods {
		commentTypes[k] = true
	}
	// 添加用于注释回退的基础类型
	commentTypes["aggregateSql"] = true

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		return result
	}

	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}

		// 获取接收者类型名
		var recvType string
		switch t := fn.Recv.List[0].Type.(type) {
		case *ast.Ident:
			recvType = t.Name
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				recvType = ident.Name
			}
		}

		if recvType == "" {
			continue
		}

		// 只处理我们关心的类型
		if !commentTypes[recvType] {
			continue
		}

		// 获取方法名
		methodName := fn.Name.Name

		// 获取注释
		if fn.Doc != nil {
			var comments []string
			for _, c := range fn.Doc.List {
				text := strings.TrimPrefix(c.Text, "//")
				text = strings.TrimPrefix(text, " ")
				comments = append(comments, text)
			}
			key := recvType + "." + methodName
			result[key] = comments
		}
	}

	return result
}

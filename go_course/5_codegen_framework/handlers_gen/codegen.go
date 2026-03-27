package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	genFuncPrefix = "apigen:api"
)

type structTpl struct {
	StructName string
	RecvName   string
}

func main() {
	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal().Err(err).Msg("parse file")
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	err = importPackagesTpl.Execute(out, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("execute import packages template")
	}

	genStructsWithMethods := make(map[string][]*ast.FuncDecl)
	structToRecvName := make(map[string]string)

	for _, decl := range node.Decls {
		fDecl, ok := decl.(*ast.FuncDecl)
		if ok {
			if fDecl.Doc == nil || !strings.HasPrefix(fDecl.Doc.Text(), genFuncPrefix) {
				// интересуют только методы с префиксом apigen:api
				continue
			}
			structName := resolveStructName(fDecl)
			genStructsWithMethods[structName] = append(genStructsWithMethods[structName], fDecl)
			structToRecvName[structName] = fDecl.Recv.List[0].Names[0].Name
		}
	}

	for structName, structMethods := range genStructsWithMethods {
		recvName := structToRecvName[structName]
		err = startStructTpl.Execute(out, structTpl{StructName: structName, RecvName: recvName})
		if err != nil {
			log.Fatal().Err(err).Msg("execute start struct template")
		}

		for _, structMethod := range structMethods {
			processServeFuncDecl(structMethod, recvName, out)
		}

		err = endStructTpl.Execute(out, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("execute end struct template")
		}

		for _, structMethod := range structMethods {
			processMethodFuncDecl(structMethod, structName, recvName, out)
		}
	}
}

func resolveStructName(decl *ast.FuncDecl) string {
	return decl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
}

type apiSpec struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`

	RecvName   string
	StructName string
	MethodName string
}

func processServeFuncDecl(fDecl *ast.FuncDecl, recvName string, out *os.File) {
	spec := parseApiSpec(fDecl.Doc.Text())
	spec.RecvName = recvName
	spec.MethodName = fDecl.Name.Name
	err := caseMethodTlp.Execute(out, spec)
	if err != nil {
		log.Fatal().Err(err).Msg("execute case method template")
	}
}

func parseApiSpec(comm string) *apiSpec {
	obj := strings.TrimPrefix(comm, genFuncPrefix)
	spec := &apiSpec{}
	err := json.Unmarshal([]byte(obj), spec)
	if err != nil {
		log.Fatal().Err(err).Msg("parse api spec")
	}
	return spec
}

func processMethodFuncDecl(fDecl *ast.FuncDecl, structName string, recvName string, out *os.File) {
	methodName := fDecl.Name.Name

	spec := parseApiSpec(fDecl.Doc.Text())
	spec.RecvName = recvName
	spec.StructName = structName
	spec.MethodName = methodName

	err := wrapperMethodStartTpl.Execute(out, spec)
	if err != nil {
		log.Fatal().Err(err).Msg("execute wrapper method start template")
	}

	if spec.Method != "" {
		err = methodCheckTpl.Execute(out, spec)
		if err != nil {
			log.Fatal().Err(err).Msg("execute method check template")
		}
	}

	if spec.Auth {
		err = authCheckTpl.Execute(out, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("execute auth check template")
		}
	}

	processMethodParams(fDecl, out)
	processMethodCall(recvName, methodName, out)
}

func processMethodParams(fDecl *ast.FuncDecl, out *os.File) {
	params, ok := fDecl.Type.Params.List[1].Type.(*ast.Ident)
	if !ok {
		log.Fatal().
			Err(errors.New("cast method argument with params to *ast.Ident")).
			Msg("on process method params")
	}
	paramsTypeSpec, ok := params.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		log.Fatal().
			Err(errors.New("cast params to *ast.TypeSpec")).
			Msg("on process method params")
	}
	paramsStructType, ok := paramsTypeSpec.Type.(*ast.StructType)
	if !ok {
		log.Fatal().
			Err(errors.New("cast params type spec to *ast.StructType")).
			Msg("on process method params")
	}

	paramsFields := paramsStructType.Fields.List
	paramsStructName := paramsTypeSpec.Name.Name

	for _, f := range paramsStructType.Fields.List {
		processParamField(f, out)
	}
	processCreateParamsStruct(paramsFields, paramsStructName, out)
}

var tagRegex = regexp.MustCompile(`([\w-]+):"([^"]*)"`)

const (
	apiValidatorTag = "apivalidator"
	intType         = "int"
)

type fieldData struct {
	FldName        string
	ParamName      string
	IsInt          bool
	Required       bool
	DefaultValue   any
	MinValue       int
	MaxValue       int
	EnumValsStr    string
	EnumValsErrStr string
}

func processParamField(f *ast.Field, out *os.File) {
	fName := f.Names[0].Name
	fTypeIdent, ok := f.Type.(*ast.Ident)
	if !ok {
		log.Fatal().
			Err(errors.New("cast field type expr to *ast.Ident")).
			Msg("on process param field")
	}
	fType := fTypeIdent.Name

	fldTag := strings.Trim(f.Tag.Value, "`")
	labelsMap := make(map[string]string)
	if strings.HasPrefix(fldTag, apiValidatorTag) {
		s := tagRegex.FindStringSubmatch(fldTag)
		labelsStr := s[2]
		labels := strings.Split(labelsStr, ",")
		for _, l := range labels {
			labelParts := strings.Split(l, "=")
			if len(labelParts) > 1 {
				labelsMap[labelParts[0]] = labelParts[1]
			} else {
				labelsMap[labelParts[0]] = ""
			}
		}
	}

	paramName := strings.ToLower(fName)
	if v, ok := labelsMap["paramname"]; ok {
		paramName = v
	}
	_, reqFld := labelsMap["required"]
	err := getFieldValTpl.Execute(out, &fieldData{
		FldName: fName, ParamName: paramName, IsInt: fType == intType, Required: reqFld,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("execute get field value template")
	}

	if defVal, hasDefault := labelsMap["default"]; hasDefault {
		err = resolveDefaultFieldValTpl.Execute(out, &fieldData{
			FldName: fName, IsInt: fType == intType, DefaultValue: defVal,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("execute resolve default field value template")
		}
	}

	if minValStr, hasMin := labelsMap["min"]; hasMin {
		minVal, _ := strconv.Atoi(minValStr)
		err = checkMinFieldValTpl.Execute(out, &fieldData{
			FldName: fName, ParamName: paramName, IsInt: fType == intType, MinValue: minVal,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("execute check min field value template")
		}
	}

	if maxValStr, hasMax := labelsMap["max"]; hasMax {
		maxVal, _ := strconv.Atoi(maxValStr)
		err = checkMaxFieldValTpl.Execute(out, &fieldData{
			FldName: fName, ParamName: paramName, IsInt: fType == intType, MaxValue: maxVal,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("execute check max field value template")
		}
	}

	if enumValsStr, hasEnum := labelsMap["enum"]; hasEnum {
		enumVals := strings.Split(enumValsStr, "|")
		err = checkEnumFieldValTpl.Execute(out, &fieldData{
			FldName:        fName,
			ParamName:      paramName,
			IsInt:          fType == intType,
			EnumValsStr:    strings.Join(enumVals, "\", \""),
			EnumValsErrStr: strings.Join(enumVals, ", "),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("execute check enum field value template")
		}
	}

	fmt.Fprintln(out)
}

type paramsStructData struct {
	ParamsStructName string
	Fields           []string
}

func processCreateParamsStruct(paramsFields []*ast.Field, paramsStructName string, out *os.File) {
	fields := make([]string, 0, len(paramsFields))
	for _, f := range paramsFields {
		fields = append(fields, f.Names[0].Name)
	}
	err := createParamsStructTpl.Execute(out, &paramsStructData{
		ParamsStructName: paramsStructName,
		Fields:           fields,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("execute create params struct template")
	}
}

type callMathodData struct {
	RecvName   string
	MethodName string
}

func processMethodCall(recvName string, methodName string, out *os.File) {
	err := callMethodTpl.Execute(out, &callMathodData{
		RecvName:   recvName,
		MethodName: methodName,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("execute call method template")
	}
}

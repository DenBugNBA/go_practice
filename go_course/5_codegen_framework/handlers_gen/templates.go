package main

import "text/template"

var (
	importPackagesTpl = template.Must(template.New("importPackagesTpl").Parse(`
import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
)
`))

	startStructTpl = template.Must(template.New("startStructTpl").Parse(`
func ({{.RecvName}} *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {`))

	caseMethodTlp = template.Must(template.New("caseMethodTpl").Parse(`
	case "{{.Url}}":
		{{.RecvName}}.wrapper{{.MethodName}}(w, r)`))

	endStructTpl = template.Must(template.New("endStructTpl").Parse(`
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown method\"}"))
	}
}
`))

	wrapperMethodStartTpl = template.Must(template.New("wrapperMethodStartTpl").Parse(`
func ({{.RecvName}} *{{.StructName}}) wrapper{{.MethodName}}(w http.ResponseWriter, r *http.Request) {`))

	methodCheckTpl = template.Must(template.New("methodCheckTpl").Parse(`
	if r.Method != "{{.Method}}" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("{\"error\": \"bad method\"}"))
		return
	}`))

	authCheckTpl = template.Must(template.New("authCheckTpl").Parse(`
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("{\"error\": \"unauthorized\"}"))
		return
	}`))

	getFieldValTpl = template.Must(template.New("getFieldValTpl").Parse(`
	rawParam{{.FldName}} := r.FormValue("{{.ParamName}}") 
	{{- if .IsInt }}
	param{{.FldName}}, err := strconv.Atoi(rawParam{{.FldName}})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"{{.ParamName}} must be int\"}"))
		return
	}
	{{- else }}
	param{{.FldName}} := rawParam{{.FldName}}
	{{- end }}
	{{- if .Required }}
		{{- if .IsInt }}
	if param{{.FldName}} == 0 {
		{{- else }}
	if param{{.FldName}} == "" {
	{{- end }}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"{{.ParamName}} must not be empty\"}"))
		return
	}
	{{- end }}`))

	resolveDefaultFieldValTpl = template.Must(template.New("resolveDefaultFieldValTpl").Parse(`
	{{- if .IsInt }}
	if param{{.FldName}} == 0 {
		param{{.FldName}} = {{ .DefaultValue }}
	{{- else }}
	if param{{.FldName}} == "" {
		param{{.FldName}} = "{{ .DefaultValue }}"
	{{- end }}
	}`))

	checkMinFieldValTpl = template.Must(template.New("checkMinFieldValTpl").Parse(`
	{{- if .IsInt }}
	if param{{.FldName}} < {{ .MinValue }}  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"{{.ParamName}} must be >= {{ .MinValue }}\"}"))
		return
	{{- else }}
	if len(param{{.FldName}}) < {{ .MinValue }}  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"{{.ParamName}} len must be >= {{ .MinValue }}\"}"))
		return
	{{- end }}
	}`))

	checkMaxFieldValTpl = template.Must(template.New("checkMaxFieldValTpl").Parse(`
	{{- if .IsInt }}
	if param{{.FldName}} > {{ .MaxValue }}  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"{{.ParamName}} must be <= {{ .MaxValue }}\"}"))
		return
	{{- else }}
	if len(param{{.FldName}}) > {{ .MaxValue }} {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"{{.ParamName}} len must be <= {{ .MaxValue }}\"}"))
		return
	{{- end }}
	}`))

	checkEnumFieldValTpl = template.Must(template.New("checkEnumFieldValTpl").Parse(`
	supportedValues := []string{"{{ .EnumValsStr }}"}
	if !slices.Contains(supportedValues, param{{.FldName}}) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"{{.ParamName}} must be one of [{{ .EnumValsErrStr}}]\"}"))
		return
	}`))

	createParamsStructTpl = template.Must(template.New("createParamsStructTpl").Parse(`
	methodParams := {{ .ParamsStructName }}{
	{{- range .Fields }}
		{{ . }}: param{{ . }},
	{{- end}}
	}`))

	callMethodTpl = template.Must(template.New("callMethodTpl").Parse(`
	res, err := {{ .RecvName }}.{{ .MethodName }}(r.Context(), methodParams)
	if err != nil {
		var apiErr ApiError
		if errors.As(err, &apiErr) {
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		}
		return
	}
	resp := make(map[string]any, 2)
	resp["response"] = res
	resp["error"] = ""
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
`))
)

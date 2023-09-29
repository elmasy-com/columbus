package frontend

import (
	"embed"
	"fmt"
	"html/template"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templatesFS embed.FS
var templates = template.Must(template.New("").Funcs(registerTemplateFuncs()).ParseFS(templatesFS, "templates/*"))

func printUnixDate(sec int64) string {

	return time.Unix(sec, 0).UTC().Format("2006-01-02 15:04")
}

func prettyPrintInt64(n int64) string {

	return message.NewPrinter(language.English).Sprint(n)
}

func prettyPrintFloat64(n float64) string {

	return fmt.Sprintf("%.2f", n)
}

func derefSubData(t *SubData) SubData {
	return *t
}

func derefSubDatas(ts []*SubData) []SubData {

	v := make([]SubData, 0, len(ts))

	for i := range ts {
		v = append(v, *ts[i])
	}
	return v
}

func registerTemplateFuncs() template.FuncMap {

	fm := make(map[string]any)

	fm["printUnixDate"] = printUnixDate
	fm["prettyPrintInt64"] = prettyPrintInt64
	fm["prettyPrintFloat64"] = prettyPrintFloat64
	fm["derefSubData"] = derefSubData
	fm["derefSubDatas"] = derefSubDatas

	return fm
}

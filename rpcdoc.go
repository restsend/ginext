package ginext

import (
	_ "embed"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
)

const ApiDocsJSONUri = "/docs/api.json"
const ApiDocsUri = "/docs/api"

type RpcDocField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type ResultType struct {
	Name    string       `json:"name"`
	Type    string       `json:"type"`
	CanNull bool         `json:"canNull"`
	Fields  []ResultType `json:"fields,omitempty"`
}

type RpcDoc struct {
	AuthRequired bool `json:"authRequired"`
	OnlyPost     bool `json:"onlyPost"`
	//Form
	Fields       []RpcDocField `json:"fields,omitempty"`
	ResultType   ResultType    `json:"resultType,omitempty"`
	RelativePath string        `json:"uri"`
	DocString    string        `json:"doc"`
	IsGroup      bool          `json:"isGroup"`
}

var rpcDocs []RpcDoc

func parseResultType(rt reflect.Type, name string) (val ResultType) {
	val.Name = name

	if rt.Kind() == reflect.Struct {
		val.Type = "object"
	} else if rt.Kind() == reflect.Array || rt.Kind() == reflect.Slice {
		val.Type = "[]object"
		rt = rt.Elem()
	} else if rt.Kind() == reflect.Ptr {
		val.Type = "object"
		val.CanNull = true
		rt = rt.Elem()
	}

	if rt.Kind() == reflect.Struct {
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			jsonTag := f.Tag.Get("json")
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				//
				embedRT := parseResultType(f.Type, "")
				val.Fields = append(val.Fields, embedRT.Fields...)
				continue
			}
			if len(jsonTag) <= 0 || jsonTag == "-" {
				continue
			}

			fieldName := strings.Split(jsonTag, ",")[0]
			if f.Type.Name() == "Time" {
				fieldRT := ResultType{
					Name: fieldName,
					Type: "Date",
				}
				val.Fields = append(val.Fields, fieldRT)
			} else if f.Type.Name() == "NullTime" {
				fieldRT := ResultType{
					Name: fieldName,
					Type: "Date",
				}
				fieldRT.CanNull = true
				val.Fields = append(val.Fields, fieldRT)
			} else {
				fieldRT := parseResultType(f.Type, fieldName)
				val.Fields = append(val.Fields, fieldRT)
			}
		}
		return val
	}

	switch rt.Kind() {
	case reflect.Bool:
		val.Type = "boolean"
	case reflect.String:
		val.Type = "string"
	case reflect.Map:
		val.Type = "{}"
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		val.Type = "Integer"
	}
	return val
}

func asJsonType(intype string) string {
	if strings.Contains(intype, "Time") {
		return "Date"
	} else if strings.Contains(intype, "int") {
		return "Integer"
	} else if strings.Contains(intype, "bool") {
		return "boolean"
	}
	return intype
}

func parseFileds(s interface{}) []RpcDocField {
	if s == nil {
		return nil
	}

	rt := reflect.TypeOf(s)
	if rt.Kind() != reflect.Struct {
		return nil
	}
	var docFields []RpcDocField
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)

		jsonTag := f.Tag.Get("json")

		if len(jsonTag) <= 0 || jsonTag == "-" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]
		docField := RpcDocField{
			Name:     fieldName,
			Required: false,
		}

		if f.Type.Kind() != reflect.Ptr {
			docField.Type = f.Type.Name()
		} else {
			docField.Type = f.Type.String()[1:]
		}

		docField.Type = asJsonType(docField.Type)

		if f.Tag.Get("binding") == "required" {
			docField.Required = true
		}
		docFields = append(docFields, docField)
	}

	return docFields
}

func AddDocAppLabel(label string) {
	doc := RpcDoc{
		IsGroup:      true,
		RelativePath: label,
	}
	rpcDocs = append(rpcDocs, doc)
}

func AddDoc(ctx *RpcContext) {
	doc := RpcDoc{
		AuthRequired: ctx.AuthRequired,
		OnlyPost:     ctx.OnlyPost,
		RelativePath: ctx.RelativePath,
		Fields:       parseFileds(ctx.Form),
	}

	if ctx.Result != nil {
		doc.ResultType = parseResultType(reflect.TypeOf(ctx.Result), "")
	}

	if len(ctx.Doc) > 0 {
		doc.DocString = string(markdown.ToHTML([]byte(ctx.Doc), nil, nil))
	}
	rpcDocs = append(rpcDocs, doc)
}

//go:embed assets/rpcdoc.html
var rpcDocHtml string

func registerDocHandler(r *gin.Engine) {
	rpcDocs = make([]RpcDoc, 0)
	r.GET(ApiDocsJSONUri, func(c *gin.Context) {
		c.JSON(http.StatusOK, rpcDocs)
	})

	r.GET(ApiDocsUri, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(rpcDocHtml))
	})
}

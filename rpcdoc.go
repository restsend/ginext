package ginext

import (
	_ "embed"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

const ApiDocsJSONUri = "/docs/api.json"
const ApiDocsUri = "/docs/api"

type RpcFieldType struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Required bool           `json:"required"`
	CanNull  bool           `json:"canNull"`
	Fields   []RpcFieldType `json:"fields,omitempty"`
}

type RpcDoc struct {
	AuthRequired bool `json:"authRequired"`
	OnlyPost     bool `json:"onlyPost"`
	//Form
	Fields       []RpcFieldType `json:"fields,omitempty"`
	ResultType   RpcFieldType   `json:"resultType,omitempty"`
	RelativePath string         `json:"uri"`
	DocString    string         `json:"doc"`
	IsGroup      bool           `json:"isGroup"`
}

var rpcDocs []RpcDoc

func parseResultType(rt reflect.Type, name string, stacks []string) (val RpcFieldType) {
	val.Name = name

	if rt.Kind() == reflect.Ptr {
		val.CanNull = true
		rt = rt.Elem()
	}
	if rt.Name() == "NullTime" {
		val.CanNull = true
		val.Type = "Date"
		return val
	} else if rt.Name() == "Time" {
		val.Type = "Date"
		return val
	} else if rt.Kind() == reflect.Struct {
		val.Type = "object"
	} else if rt.Kind() == reflect.Array || rt.Kind() == reflect.Slice {
		val.Type = "[]object"
		rt = rt.Elem()
	} else if rt.Kind() == reflect.Ptr {
		val.Type = "object"
	}

	if rt.Kind() == reflect.Struct {
		for _, v := range stacks {
			if rt.Name() == v {
				val.Type = "object"
				return val
			}
		}

		stacks = append(stacks, rt.Name())
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			jsonTag := f.Tag.Get("json")
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				//
				embedRT := parseResultType(f.Type, "", stacks)
				val.Fields = append(val.Fields, embedRT.Fields...)
				continue
			}
			if len(jsonTag) <= 0 || jsonTag == "-" {
				continue
			}

			fieldName := strings.Split(jsonTag, ",")[0]
			if f.Type.Name() == "Time" {
				fieldRT := RpcFieldType{
					Name: fieldName,
					Type: "Date",
				}
				val.Fields = append(val.Fields, fieldRT)
			} else if f.Type.Name() == "NullTime" {
				fieldRT := RpcFieldType{
					Name: fieldName,
					Type: "Date",
				}
				fieldRT.CanNull = true
				val.Fields = append(val.Fields, fieldRT)
			} else {
				fieldRT := parseResultType(f.Type, fieldName, stacks)
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

func parseFileds(rt reflect.Type) []RpcFieldType {
	if rt.Kind() != reflect.Struct {
		return nil
	}
	var docFields []RpcFieldType
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			//
			embedRT := parseFileds(f.Type)
			docFields = append(docFields, embedRT...)
			continue
		}

		jsonTag := f.Tag.Get("json")

		if len(jsonTag) <= 0 || jsonTag == "-" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]
		docField := parseResultType(f.Type, fieldName, nil)
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
	}
	if ctx.Form != nil {
		doc.Fields = parseFileds(reflect.TypeOf(ctx.Form))
	}

	if ctx.Result != nil {
		doc.ResultType = parseResultType(reflect.TypeOf(ctx.Result), "", nil)
	}

	if len(ctx.Doc) > 0 {
		doc.DocString = ctx.Doc
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

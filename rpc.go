package ginext

import (
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type RpcContext struct {
	AuthRequired bool
	OnlyPost     bool
	Form         interface{}
	Result       interface{}
	RelativePath string
	Handler      gin.HandlerFunc
	//Markdown Document
	Doc string
}

func RpcOk(c *gin.Context, obj interface{}) {
	resultType, ok := c.Get(RpcResultField)
	if ok && resultType != nil {
		if reflect.TypeOf(resultType) != reflect.TypeOf(obj) {
			log.Printf("incorrect result type %s required:%s != result:%s", c.Request.URL.Path, reflect.TypeOf(resultType), reflect.TypeOf(obj))
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": obj,
	})
}

func RpcFail(c *gin.Context, failCode int, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": failCode,
		"msg":  msg,
	})
}

func RpcError(c *gin.Context, err error) {
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusBadRequest,
		"msg":  err.Error(),
	})
}

func RpcDefine(r *gin.Engine, ctx *RpcContext) {
	funcObj := func(c *gin.Context) {
		c.Set(RpcResultField, ctx.Result)
		//if ctx.AuthRequired {
		//}
		// Process Token
		//
		if ctx.Form != nil {
			// Init Form object
			form := reflect.New(reflect.TypeOf(ctx.Form)).Interface()
			var err error

			if c.Request.Method == "POST" && c.Request.ContentLength > 0 {
				err = c.BindJSON(&form)
			} else if c.Request.Method == "GET" {
				err = c.ShouldBindQuery(form)
			}

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"type":  "json bind",
					"error": err.Error(),
				})
				return
			}

			//If form decode fail, the `form` will be nil.
			if form == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid requst body",
				})
				return
			}

			c.Set(RpcFormField, form)
		}
		ctx.Handler(c)
	}

	r.POST(ctx.RelativePath, funcObj)
	if !ctx.OnlyPost {
		r.GET(ctx.RelativePath, funcObj)
	}
	AddDoc(ctx)
}
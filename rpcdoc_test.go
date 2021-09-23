package ginext

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testEmbedResult struct {
	Foo      string       `json:"foo"`
	DeleteAt sql.NullTime `json:"deleteAt"`
	UserInfoResult
}

func TestDocString(t *testing.T) {
	cfg := NewGinExt("..")
	cfg.Init()

	r := gin.Default()
	cfg.WithGinExt(r)

	RpcDefine(r, &RpcContext{
		Form:         RegisterUserForm{},
		Result:       UserInfoResult{},
		OnlyPost:     true,
		RelativePath: "/mockapi/register",
		Handler: func(c *gin.Context) {
			RpcOk(c, UserInfoResult{
				UserName: "mockuser",
				Email:    "mockemail",
			})
		},
	})

	RpcDefine(r, &RpcContext{
		Form:         RegisterUserForm{},
		Result:       testEmbedResult{},
		OnlyPost:     true,
		RelativePath: "/mockapi/embed",
		Handler: func(c *gin.Context) {
			RpcOk(c, UserInfoResult{
				UserName: "mockuser",
				Email:    "mockemail",
			})
		},
	})

	client := NewTestHTTPClient(r)

	w := client.Get(ApiDocsJSONUri)
	assert.Equal(t, http.StatusOK, w.Code)

	var rpcDocs []RpcDoc
	json.Unmarshal(w.Body.Bytes(), &rpcDocs)

	assert.NotNil(t, rpcDocs)
	assert.Equal(t, len(rpcDocs), 2)
	assert.Equal(t, len(rpcDocs[0].Fields), 8)
	assert.Equal(t, rpcDocs[0].Fields[0].Name, "email")
	assert.Equal(t, rpcDocs[0].Fields[0].Type, "string")
	assert.True(t, rpcDocs[0].Fields[0].Required)

	assert.Equal(t, rpcDocs[0].ResultType.Type, "object")
	log.Println(rpcDocs[1].ResultType)
	assert.Equal(t, len(rpcDocs[1].ResultType.Fields), 5)
	assert.Equal(t, rpcDocs[1].ResultType.Fields[0].Name, "foo")
	assert.Equal(t, rpcDocs[1].ResultType.Fields[0].Type, "string")
	assert.Equal(t, rpcDocs[1].ResultType.Fields[1].Name, "deleteAt")
	assert.Equal(t, rpcDocs[1].ResultType.Fields[1].Type, "Date")
	assert.Equal(t, rpcDocs[1].ResultType.Fields[2].Name, "username")
	assert.Equal(t, rpcDocs[1].ResultType.Fields[2].Type, "string")

	w = client.Get(ApiDocsUri)
	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, ApiDocsJSONUri)
}

func TestParseField(t *testing.T) {
	type testForm struct {
		PaginationForm
		Val string `json:"val"`
	}
	f := testForm{}
	fields := parseFileds(reflect.TypeOf(f))
	assert.Equal(t, 5, len(fields))
	assert.Equal(t, "keyword", fields[0].Name)
	assert.Equal(t, "val", fields[4].Name)
}

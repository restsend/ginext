package ginext

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestHTTPClient struct {
	r         http.Handler
	cookieJar http.CookieJar
	Scheme    string
	Host      string
}

func NewTestHTTPClient(r http.Handler) (c *TestHTTPClient) {
	jar, _ := cookiejar.New(nil)
	return &TestHTTPClient{
		r:         r,
		cookieJar: jar,
		Scheme:    "http",
		Host:      "1.2.3.4",
	}
}

func (c *TestHTTPClient) SendReq(path string, req *http.Request) *httptest.ResponseRecorder {
	req.URL.Scheme = "http"
	req.URL.Host = "MOCKSERVER"

	currentUrl := &url.URL{
		Scheme: c.Scheme,
		Host:   c.Host,
		Path:   path,
	}

	cookies := c.cookieJar.Cookies(currentUrl)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	c.r.ServeHTTP(w, req)
	c.cookieJar.SetCookies(currentUrl, w.Result().Cookies())
	return w
}

//TestDoGet Quick Test Get
func (c *TestHTTPClient) Get(path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", path, nil)
	return c.SendReq(path, req)
}

//TestDoGet Quick Test Post
func (c *TestHTTPClient) Post(path string, param map[string]interface{}) *httptest.ResponseRecorder {
	body, _ := json.Marshal(param)
	return c.PostRaw(path, body)
}

func (c *TestHTTPClient) PostRaw(path string, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	return c.SendReq(path, req)
}

//TestDoGet Quick Test CheckResponse
func (c *TestHTTPClient) CheckResponse(t *testing.T, w *httptest.ResponseRecorder) (response map[string]interface{}) {
	assert.Equal(t, http.StatusOK, w.Code)
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	return response
}

func CheckSubSet(t *testing.T, expected, actual map[string]interface{}) {
	for k, v := range expected {
		av, ok := actual[k]
		assert.True(t, ok, "not found:"+k)
		if v == nil {
			assert.Nil(t, av)
			continue
		}
		vt := reflect.TypeOf(v)
		avt := reflect.ValueOf(av)
		var newValue reflect.Value
		{
			defer func() {
				if recover() != nil {
					assert.Fail(t, "value type not equal, expect type:", vt, ", actual type:", avt)
				}
			}()
			newValue = avt.Convert(vt)
		}
		assert.Equal(t, v, newValue.Interface(), "not equal:"+k)
	}
}

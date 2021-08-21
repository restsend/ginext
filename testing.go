package ginext

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
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

// Rpc Call
func (c *TestHTTPClient) Call(path string, form interface{}, result interface{}) error {
	body, err := json.Marshal(form)
	if err != nil {
		return err
	}
	w := c.PostRaw(path, body)
	if w.Code != http.StatusOK {
		return errors.New("bad status :" + w.Result().Status)
	}
	respData := map[string]interface{}{}
	err = json.Unmarshal(w.Body.Bytes(), &respData)
	if err != nil {
		return err
	}
	code := int(respData["code"].(float64))
	if code != http.StatusOK {
		return errors.New(respData["msg"].(string))
	}

	data, ok := respData["data"]
	if !ok {
		return errors.New("bad resp not data key")
	}

	if result == nil {
		return nil
	}
	content, _ := json.Marshal(data)
	return json.Unmarshal(content, result)
}

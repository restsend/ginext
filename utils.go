package ginext

import (
	"bytes"
	"errors"
	"log"
	"math/rand"
	"reflect"
	"text/template"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
var numberRunes = []rune("0123456789")

func randRunes(n int, source []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = source[rand.Intn(len(source))]
	}
	return string(b)
}

func RandText(n int) string {
	return randRunes(n, letterRunes)
}

func RandNumberText(n int) string {
	return randRunes(n, numberRunes)
}

func FormatData(fmtText string, ctx map[string]string, fallbackFmt string) string {
	tmplRoot := template.New("ginext render")
	tmpl, err := tmplRoot.Parse(fmtText)
	if len(fmtText) <= 0 || err != nil {
		tmpl, err = tmplRoot.Parse(fallbackFmt)
		if err != nil {
			return ""
		}
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, ctx)
	if err != nil {
		return ""
	}
	return b.String()
}

func SafeCall(f func() error, failHandle func(error)) error {
	defer func() {
		if err := recover(); err != nil {
			if failHandle != nil {
				eo, ok := err.(error)
				if !ok {
					es, ok := err.(string)
					if ok {
						eo = errors.New(es)
					} else {
						eo = errors.New("unknown error type")
					}
				}
				failHandle(eo)
			} else {
				log.Println(err)
			}
		}
	}()
	return f()
}

func FormAsMap(form interface{}, fields []string) (vals map[string]interface{}) {
	vals = make(map[string]interface{})
	v := reflect.ValueOf(form)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return vals
	}
	for i := 0; i < len(fields); i++ {
		k := v.FieldByName(fields[i])
		if !k.IsValid() || k.IsZero() {
			continue
		}
		if k.Kind() == reflect.Ptr {
			if !k.IsNil() {
				vals[fields[i]] = k.Elem().Interface()
			}
		} else {
			vals[fields[i]] = k.Interface()
		}
	}
	return vals
}

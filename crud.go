package ginext

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListObject(c *gin.Context, tx *gorm.DB, r interface{}, form *PaginationForm, order, searchKey string) {
	ListObjectEx(c, tx, r, form, order, searchKey, nil)
}

func ListObjectEx(c *gin.Context, tx *gorm.DB, r interface{}, form *PaginationForm, order, searchKey string, resultCallback func(r interface{})) {

	if c == nil {
		return
	}
	if form != nil {
		if len(form.GetKeyword()) > 0 {
			sc := strings.Count(searchKey, "?")
			k := "%" + form.GetKeyword() + "%"
			args := make([]interface{}, sc)
			for i := 0; i < sc; i++ {
				args = append(args, k)
			}
			tx = tx.Where(searchKey, args...)
		}
	}

	var tc int64
	result := tx.Count(&tc)

	if result.Error != nil {
		RpcError(c, result.Error)
		return
	}

	rv := reflect.ValueOf(r)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rv.FieldByName("TotalCount").SetInt(tc)
	items := rv.FieldByName("Items").Addr().Interface()

	if form != nil {
		tx = tx.Offset(form.GetPos()).Limit(form.GetLimit())
	}

	result = tx.Find(items)

	if result.Error != nil {
		RpcError(c, result.Error)
		return
	}

	if form != nil {
		ic := rv.FieldByName("Items").Len()
		pos := form.GetPos() + ic
		rv.FieldByName("Pos").SetInt(int64(pos))
		rv.FieldByName("Limit").SetInt(int64(form.GetLimit()))
	}

	if resultCallback != nil {
		resultCallback(r)
	}
	RpcOk(c, rv.Interface())
}

func NewObject(c *gin.Context, db *gorm.DB, modPtr interface{}) {
	result := db.Create(modPtr)
	if result.Error != nil {
		if c != nil {
			RpcError(c, result.Error)
		}
		return
	}
	if c != nil {
		RpcOk(c, reflect.ValueOf(modPtr).Elem().Interface())
	}
}

func DeleteObject(c *gin.Context, db *gorm.DB, modPtr interface{}, ID uint, markDelete bool) {
	var result *gorm.DB
	tx := db.Model(modPtr).Where("id", ID)
	if markDelete {
		result = tx.UpdateColumn("Deleted", true)
	} else {
		result = tx.Delete(modPtr)
	}

	if c == nil {
		return
	}

	if result.Error != nil {
		RpcError(c, result.Error)
		return
	}

	if result.RowsAffected <= 0 {
		RpcOk(c, false)
	} else {
		RpcOk(c, true)
	}
}

func EditObject(c *gin.Context, db *gorm.DB, modPtr interface{}, ID uint, vals map[string]interface{}) {
	result := db.Model(modPtr).Where("id", ID).Updates(vals)
	if c == nil {
		return
	}

	if result.Error != nil {
		RpcError(c, result.Error)
		return
	}
	result = db.Model(modPtr).Take(modPtr, ID)
	if result.Error != nil {
		RpcError(c, result.Error)
		return
	}

	RpcOk(c, reflect.ValueOf(modPtr).Elem().Interface())
}

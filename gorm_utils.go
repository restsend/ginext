package ginext

import "gorm.io/gorm"

func GenUniqueKey(tx *gorm.DB, field string, size int) (key string) {
	key = RandText(size)
	for i := 0; i < 10; i++ {
		var c int64
		result := tx.Where(field, key).Limit(1).Count(&c)
		if result.Error != nil {
			break
		}
		if c > 0 {
			continue
		}
		return key
	}
	return "BAD:" + key
}

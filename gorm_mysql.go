// +build mysql

package ginext

import (
	"errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func (c *GinExt) createDatabaseInstance(cfg *gorm.Config) error {
	var err error
	if c.DbDriver == "mysql" {
		c.DbInstance, err = gorm.Open(mysql.Open(c.DbDSN), cfg)
	} else {
		err = errors.New("unknown driver" + c.DbDriver)
	}
	return err
}

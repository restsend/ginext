// +build !mysql

package ginext

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (c *GinExt) createDatabaseInstance(cfg *gorm.Config) error {
	var err error
	if c.DbDriver == "sqlite" {
		c.DbInstance, err = gorm.Open(sqlite.Open(c.DbDSN), cfg)
	} else if c.DbDriver == "mysql" {
		c.DbInstance, err = gorm.Open(mysql.Open(c.DbDSN), cfg)
	}
	return err
}

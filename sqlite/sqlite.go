package sqlite

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var gormDB *gorm.DB

func SetSqlite() {
	// home, err := os.UserHomeDir()
	// if err != nil {
	// 	log.Fatalf("创建本地sqlite数据库目录失败:%s", err.Error())
	// }
	// location := filepath.Join(home, "duplicate.db")
	location := "duplicate.db"
	db, err := gorm.Open(sqlite.Open(location), &gorm.Config{})
	if err != nil {
		log.Fatalf("打开本地sqlite数据库失败:%s", err.Error())
	}

	// 自动建表
	err = db.AutoMigrate(&FileHash{})
	if err != nil {
		log.Fatalf("创建数据表失败:%s", err.Error())
	}

	gormDB = db
	log.Println("本地sqlite数据库初始化完成")
}

func GetSqlite() *gorm.DB {
	return gormDB
}

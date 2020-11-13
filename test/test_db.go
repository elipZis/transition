package test

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

// TestDB initialize a db for testing
func TestDB() *gorm.DB {
	var db *gorm.DB
	var err error
	var dbuser, dbpwd, dbname, dbhost = "elipzis", "elipzis", "elipzis_test", "localhost"

	if os.Getenv("DB_USER") != "" {
		dbuser = os.Getenv("DB_USER")
	}

	if os.Getenv("DB_PWD") != "" {
		dbpwd = os.Getenv("DB_PWD")
	}

	if os.Getenv("DB_NAME") != "" {
		dbname = os.Getenv("DB_NAME")
	}

	if os.Getenv("DB_HOST") != "" {
		dbhost = os.Getenv("DB_HOST")
	}

	loggerMode := logger.Error
	if os.Getenv("DEBUG") != "" {
		loggerMode = logger.Info
	}
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(loggerMode),
	}

	if os.Getenv("TEST_DB") == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", dbhost, dbuser, dbpwd, dbname)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	} else if os.Getenv("TEST_DB") == "mysql" {
		// CREATE USER 'elipzis'@'localhost' IDENTIFIED BY 'elipzis';
		// CREATE DATABASE elipzis_test;
		// GRANT ALL ON elipzis_test.* TO 'elipzis'@'localhost';
		dsn := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd, dbname)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	} else {
		db, err = gorm.Open(sqlite.Open("transition.db"), gormConfig)
	}

	if err != nil {
		panic(err)
	}

	return db
}

// ResetDBTables reset given tables.
func ResetDBTables(db *gorm.DB, tables ...interface{}) {
	Truncate(db, tables...)
	AutoMigrate(db, tables...)
}

// Truncate receives table arguments and truncate their content in database.
func Truncate(db *gorm.DB, givenTables ...interface{}) {
	// We need to iterate throught the list in reverse order of
	// creation, since later tables may have constraints or
	// dependencies on earlier tables.
	len := len(givenTables)
	for i := range givenTables {
		table := givenTables[len-i-1]
		db.Migrator().DropTable(table)
	}
}

// AutoMigrate receives table arguments and create or update their
// table structure in database.
func AutoMigrate(db *gorm.DB, givenTables ...interface{}) {
	for _, table := range givenTables {
		db.AutoMigrate(table)
		if migratable, ok := table.(Migratable); ok {
			exec(func() error { return migratable.AfterMigrate(db) })
		}
	}
}

// Migratable defines interface for implementing post-migration
// actions such as adding constraints that arent's supported by Gorm's
// struct tags. This function must be idempotent, since it will most
// likely be executed multiple times.
type Migratable interface {
	AfterMigrate(db *gorm.DB) error
}

func exec(c func() error) {
	if err := c(); err != nil {
		panic(err)
	}
}

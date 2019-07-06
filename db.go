package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	db *sqlx.DB
)

// Area 地区
type Area struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Parent int    `db:"parent"`
}

// InitDb 初始化数据库
func InitDb() {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Postgres.DBUser,
		config.Postgres.DBPassword,
		config.Postgres.DBHost,
		config.Postgres.DBPort,
		config.Postgres.DBName)
	db = sqlx.MustConnect("postgres", connStr)
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
}

// InsertArea 插入地区
func InsertArea(data *Area) error {
	sql := `INSERT INTO area (
						id, name, parent)
					VALUES ($1, $2, $3)`
	if _, err := db.Exec(sql, data.ID, data.Name, data.Parent); err != nil {
		fmt.Println("InsertArea出错:", err)
		return err
	}
	return nil
}

package main

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

func CreateSchema(db *pg.DB) error {
	for _, model := range []interface{}{&Player{}} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

package main

import (
	"fmt"
	"os"

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

//ConnectDataBase Подключение к базе данных
func ConnectDataBase() {
	opt, err := pg.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	db = pg.Connect(opt)

	err = CreateSchema(db)
	if err != nil {
		panic(err)
	}
	fmt.Println("Успешное подключение к базе данных")
}

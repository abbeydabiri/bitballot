package database

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"bitballot/config"
)

//Users ...
type Users struct {
	Fields
	FailedMax, Failed uint64
	Email, Mobile, Image,
	Username, Password string
}

func (table *Users) Setup() {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("bitballot"), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Setting Up Users")

	table.Password = base64.StdEncoding.EncodeToString(passwordHash)
	table.Title = "Bit Ballot"
	table.Email = "root@localhost"
	table.Workflow = "enabled"
	table.Username = "root"
	table.Create(table.ToMap())
}

func (table *Users) ToMap() (mapInterface map[string]interface{}) {
	jsonTable, _ := json.Marshal(table)
	json.Unmarshal(jsonTable, &mapInterface)
	return
}

func (table *Users) FillStruct(tableMap map[string]interface{}) error {
	jsonTable, _ := json.Marshal(tableMap)
	if err := json.Unmarshal(jsonTable, &table); err != nil {
		return err
	}
	return nil
}

func (table *Users) Create(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlInsert(table, tableMap); sqlQuery != "" {
		if err := config.Get().Postgres.Get(&table.ID, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Users) Update(tableMap map[string]interface{}) {
	if tableMap["Password"] != nil && tableMap["Password"].(string) == "" {
		delete(tableMap, "Password")
	}

	if sqlQuery, sqlParams := table.sqlUpdate(table, tableMap); sqlQuery != "" {
		if _, err := config.Get().Postgres.Exec(sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Users) GetByID(tableMap map[string]interface{}, searchParams *SearchParams) {
	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {
		sqlParams = append(sqlParams, searchParams.ID)
		sqlQuery += fmt.Sprintf("id = $%v ", len(sqlParams))
		if err := config.Get().Postgres.Get(table, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Users) Search(tableMap map[string]interface{}, searchParams *SearchParams) (list []Users) {
	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {
		searchParams.Text = "%" + searchParams.Text + "%"
		sqlParams = append(sqlParams, searchParams.Text)
		sqlQuery += fmt.Sprintf("%v like $%v order by id desc ", searchParams.Field, len(sqlParams))

		if searchParams.Limit == 0 {
			searchParams.Limit = 100
		}
		sqlParams = append(sqlParams, searchParams.Limit)
		sqlQuery += fmt.Sprintf("limit $%v ", len(sqlParams))

		sqlParams = append(sqlParams, searchParams.Skip)
		sqlQuery += fmt.Sprintf("offset $%v ", len(sqlParams))
		if err := config.Get().Postgres.Select(&list, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
	return
}

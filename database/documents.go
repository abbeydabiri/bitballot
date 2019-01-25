package database

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"bitballot/config"
)

type Documents struct {
	Fields

	Doctype, Filename, Filemeta, Filetype,
	Filepath, Validfrom, Validuntil, Dateissued,
	Issuedby, Issuedto, TableRef string

	Position int
	Filesize, OwnerID,
	TableRefID uint64
}

func (table *Documents) ToMap() (mapInterface map[string]interface{}) {
	jsonTable, _ := json.Marshal(table)
	json.Unmarshal(jsonTable, &mapInterface)
	return
}

func (table *Documents) FillStruct(tableMap map[string]interface{}) error {
	jsonTable, _ := json.Marshal(tableMap)
	if err := json.Unmarshal(jsonTable, &table); err != nil {
		return err
	}
	return nil
}

func (table *Documents) Create(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlInsert(table, tableMap); sqlQuery != "" {
		if err := config.Get().Postgres.Get(&table.ID, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Documents) Update(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlUpdate(table, tableMap); sqlQuery != "" {
		if _, err := config.Get().Postgres.Exec(sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Documents) GetByID(tableMap map[string]interface{}, searchParams *SearchParams) {
	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {
		sqlParams = append(sqlParams, searchParams.ID)
		sqlQuery += fmt.Sprintf("id = $%v ", len(sqlParams))
		if err := config.Get().Postgres.Get(table, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Documents) Search(tableMap map[string]interface{}, searchParams *SearchParams) (list []Documents) {

	if owner := searchParams.Filter["owner"]; owner != "" {
		delete(searchParams.Filter, "owner")
		searchParams.Filter["profiles.title"] = owner
	}

	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {

		reflectType := reflect.TypeOf(table).Elem()
		tablename := strings.ToLower(reflectType.Name())
		viewSearch := fmt.Sprintf("from %s where ", tablename)

		viewJoin := " from " + tablename
		viewJoin += " left join profiles on documents.ownerid = profiles.id where "

		sqlQuery = strings.Replace(sqlQuery, viewSearch, viewJoin, 1)

		if searchParams.RefID > 0 && searchParams.RefField != "" {
			sqlParams = append(sqlParams, searchParams.RefID)
			sqlQuery += fmt.Sprintf("%v.%v = $%v and ", tablename, searchParams.RefField, len(sqlParams))
		}

		searchParams.Text = "%" + searchParams.Text + "%"
		sqlParams = append(sqlParams, searchParams.Text)
		sqlQuery += fmt.Sprintf("%v.%v like $%v order by id desc ", tablename, searchParams.Field, len(sqlParams))

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

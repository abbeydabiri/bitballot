package database

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"bitballot/config"
)

type Settings struct {
	Fields
	Code string
	
	OwnerID, PartnerID uint64
}

func (table *Settings) ToMap() (mapInterface map[string]interface{}) {
	jsonTable, _ := json.Marshal(table)
	json.Unmarshal(jsonTable, &mapInterface)
	return
}

func (table *Settings) FillStruct(tableMap map[string]interface{}) error {
	jsonTable, _ := json.Marshal(tableMap)
	if err := json.Unmarshal(jsonTable, &table); err != nil {
		return err
	}
	return nil
}

func (table *Settings) Create(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlInsert(table, tableMap); sqlQuery != "" {
		if err := config.Get().Postgres.Get(&table.ID, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Settings) Update(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlUpdate(table, tableMap); sqlQuery != "" {
		if _, err := config.Get().Postgres.Exec(sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Settings) GetByID(tableMap map[string]interface{}, searchParams *SearchParams) {
	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {
		sqlParams = append(sqlParams, searchParams.ID)
		sqlQuery += fmt.Sprintf("id = $%v ", len(sqlParams))
		if err := config.Get().Postgres.Get(table, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Settings) Search(tableMap map[string]interface{}, searchParams *SearchParams) (list []Settings) {

	if owner := searchParams.Filter["owner"]; owner != "" {
		delete(searchParams.Filter, "owner")
		searchParams.Filter["owners.fullname"] = owner
	}

	if partner := searchParams.Filter["partner"]; partner != "" {
		delete(searchParams.Filter, "partner")
		searchParams.Filter["partners.fullname"] = partner
	}

	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {

		reflectType := reflect.TypeOf(table).Elem()
		tablename := strings.ToLower(reflectType.Name())
		viewSearch := fmt.Sprintf("from %s where ", tablename)

		viewJoin := " from " + tablename
		viewJoin += " left join profiles as owners on settings.ownerid = owners.id  "
		viewJoin += " left join profiles as partners on settings.partnerid = partners.id where "

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

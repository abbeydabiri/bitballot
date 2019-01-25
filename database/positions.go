package database

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"bitballot/config"
)

type Positions struct {
	Fields

	ProposalID   uint64
	MaxCandidate int
}

func (table *Positions) ToMap() (mapInterface map[string]interface{}) {
	jsonTable, _ := json.Marshal(table)
	json.Unmarshal(jsonTable, &mapInterface)
	return
}

func (table *Positions) FillStruct(tableMap map[string]interface{}) error {
	jsonTable, _ := json.Marshal(tableMap)
	if err := json.Unmarshal(jsonTable, &table); err != nil {
		return err
	}
	return nil
}

func (table *Positions) Create(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlInsert(table, tableMap); sqlQuery != "" {
		if err := config.Get().Postgres.Get(&table.ID, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Positions) Update(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlUpdate(table, tableMap); sqlQuery != "" {
		if _, err := config.Get().Postgres.Exec(sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Positions) GetByID(tableMap map[string]interface{}, searchParams *SearchParams) {
	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {
		sqlParams = append(sqlParams, searchParams.ID)
		sqlQuery += fmt.Sprintf("id = $%v ", len(sqlParams))
		if err := config.Get().Postgres.Get(table, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Positions) Search(tableMap map[string]interface{}, searchParams *SearchParams) (list []Positions) {

	if Proposal := searchParams.Filter["Proposal"]; Proposal != "" {
		delete(searchParams.Filter, "Proposal")
		searchParams.Filter["Proposals.title"] = Proposal
	}

	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {

		reflectType := reflect.TypeOf(table).Elem()
		tablename := strings.ToLower(reflectType.Name())
		viewSearch := fmt.Sprintf("from %s where ", tablename)

		viewJoin := " from " + tablename
		viewJoin += " left join proposals on positions.proposalid = Proposals.id where "

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

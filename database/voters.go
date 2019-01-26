package database

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"bitballot/config"
)

//Voters stores voters
//before being posted to blockchain
type Voters struct {
	Fields
	VoterID, PositionID, ProposalID uint64
}

func (table *Voters) ToMap() (mapInterface map[string]interface{}) {
	jsonTable, _ := json.Marshal(table)
	json.Unmarshal(jsonTable, &mapInterface)
	return
}

func (table *Voters) FillStruct(tableMap map[string]interface{}) error {
	jsonTable, _ := json.Marshal(tableMap)
	if err := json.Unmarshal(jsonTable, &table); err != nil {
		return err
	}
	return nil
}

func (table *Voters) Create(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlInsert(table, tableMap); sqlQuery != "" {
		if err := config.Get().Postgres.Get(&table.ID, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Voters) Update(tableMap map[string]interface{}) {
	if sqlQuery, sqlParams := table.sqlUpdate(table, tableMap); sqlQuery != "" {
		if _, err := config.Get().Postgres.Exec(sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Voters) GetByID(tableMap map[string]interface{}, searchParams *SearchParams) {
	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {
		sqlParams = append(sqlParams, searchParams.ID)
		sqlQuery += fmt.Sprintf("id = $%v ", len(sqlParams))
		if err := config.Get().Postgres.Get(table, sqlQuery, sqlParams...); err != nil {
			log.Println(err.Error())
		}
	}
}

func (table *Voters) Search(tableMap map[string]interface{}, searchParams *SearchParams) (list []Voters) {

	if profile := searchParams.Filter["profile"]; profile != "" {
		delete(searchParams.Filter, "profile")
		searchParams.Filter["profiles.fullname"] = profile
	}

	if position := searchParams.Filter["position"]; position != "" {
		delete(searchParams.Filter, "position")
		searchParams.Filter["positions.title"] = position
	}

	if proposal := searchParams.Filter["proposal"]; proposal != "" {
		delete(searchParams.Filter, "proposal")
		searchParams.Filter["proposals.proposal"] = proposal
	}

	if sqlQuery, sqlParams := table.sqlSelect(table, tableMap, searchParams); sqlQuery != "" {

		reflectType := reflect.TypeOf(table).Elem()
		tablename := strings.ToLower(reflectType.Name())
		viewSearch := fmt.Sprintf("from %s where ", tablename)

		viewJoin := " from " + tablename
		viewJoin += " left join profiles on voters.voterid = profiles.id  "
		viewJoin += " left join positions on voters.positionid = positions.id  "
		viewJoin += " left join proposals on voters.proposalid = proposals.id where "

		sqlQuery = strings.Replace(sqlQuery, viewSearch, viewJoin, 1)

		if searchParams.RefID > 0 && searchParams.RefField != "" {
			sqlParams = append(sqlParams, searchParams.RefID)
			sqlQuery += fmt.Sprintf("%v.%v = $%v and ", tablename, searchParams.RefField, len(sqlParams))
		}

		searchParams.Text = "%" + searchParams.Text + "%"
		sqlParams = append(sqlParams, searchParams.Text)
		sqlQuery += fmt.Sprintf("%v.%v like $%v order by profiles.lastname, profiles.firstname, positions.title, proposals.title  ", tablename, searchParams.Field, len(sqlParams))

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

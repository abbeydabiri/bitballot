package api

import (
	"encoding/json"
	"net/http"

	"github.com/justinas/alice"

	"bitballot/config"
	"bitballot/database"
)

func apiHandlerVoters(middlewares alice.Chain, router *Router) {
	router.Get("/api/voters", middlewares.ThenFunc(apiVotersGet))
	router.Post("/api/voters", middlewares.ThenFunc(apiVotersPost))
	router.Post("/api/voters/search", middlewares.ThenFunc(apiVotersSearch))
}

func apiVotersGet(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Voters{}
		table.GetByID(table.ToMap(), formSearch)

		tableMap := table.ToMap()

		Proposal := ""
		config.Get().Postgres.Get(&Proposal, "select title from proposals where id = $1 limit 1", table.ProposalID)
		tableMap["Proposal"] = Proposal

		Position := ""
		config.Get().Postgres.Get(&Position, "select title from positions where id = $1 limit 1", table.PositionID)
		tableMap["Position"] = Position

		Voter := ""
		config.Get().Postgres.Get(&Voter, "select fullname from profiles where id = $1 limit 1", table.VoterID)
		tableMap["Voter"] = Voter
		
		message.Body = tableMap
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiVotersPost(httpRes http.ResponseWriter, httpReq *http.Request) {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Voters{}
		table.FillStruct(tableMap)

		if table.Workflow == "" {
			message.Message += "Status is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		if table.VoterID == uint64(0) {
			message.Message += "Voter is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		if table.PositionID == uint64(0) {
			message.Message += "Position is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		ProposalID := uint64(0)
		config.Get().Postgres.Get(&ProposalID, "select proposalid from positions where id = $1 limit 1", table.PositionID)
		table.ProposalID = ProposalID

		if table.ProposalID == uint64(0) {
			message.Message += "Proposal is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		countID := 0
		sqlCount := "select count(id) from voters where voterid = $1 and positionid = $2 and proposalid = $3"
		config.Get().Postgres.Get(&countID, sqlCount, table.VoterID, table.PositionID, table.ProposalID)


		if countID > 0 {
			message.Message += "Voter already registered!! \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		if table.ID == 0 {
			table.FillStruct(tableMap)
			table.Create(table.ToMap())
		} else {
			table.Update(tableMap)
		}
		message.Body = table.ID
		message.Message = RecordSaved
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiVotersSearch(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureSearch(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Voters{}
		if formSearch.Field == "" {
			formSearch.Field = "Title"
		}
		var searchList []interface{}
		searchResults := table.Search(table.ToMap(), formSearch)
		for _, result := range searchResults {
			tableMap := result.ToMap()

			Proposal := ""
			config.Get().Postgres.Get(&Proposal, "select title from proposals where id = $1 limit 1", result.ProposalID)
			tableMap["Proposal"] = Proposal

			Position := ""
			config.Get().Postgres.Get(&Position, "select title from positions where id = $1 limit 1", result.PositionID)
			tableMap["Position"] = Position

			Voter := ""
			config.Get().Postgres.Get(&Voter, "select fullname from profiles where id = $1 limit 1", result.VoterID)
			tableMap["Voter"] = Voter

			searchList = append(searchList, tableMap)
		}
		message.Body = searchList

	}
	json.NewEncoder(httpRes).Encode(message)
}

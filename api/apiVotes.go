package api

import (
	"encoding/json"
	"net/http"

	"github.com/justinas/alice"

	"bitballot/config"
	"bitballot/database"
)

func apiHandlerVotes(middlewares alice.Chain, router *Router) {
	router.Get("/api/votes", middlewares.ThenFunc(apiVotesGet))
	router.Post("/api/votes", middlewares.ThenFunc(apiVotesPost))
	router.Post("/api/votes/search", middlewares.ThenFunc(apiVotesSearch))
}

func apiVotesGet(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Votes{}
		table.GetByID(table.ToMap(), formSearch)

		tableMap := table.ToMap()

		Proposal := ""
		config.Get().Postgres.Get(&Proposal, "select title from proposals where id = $1 limit 1", table.ProposalID)
		tableMap["Proposal"] = Proposal

		Position := ""
		config.Get().Postgres.Get(&Position, "select title from positions where id = $1 limit 1", table.PositionID)
		tableMap["Position"] = Position

		Candidate := ""
		config.Get().Postgres.Get(&Candidate, "select fullname from profiles where id = $1 limit 1", table.CandidateID)
		tableMap["Candidate"] = Candidate

		Voter := ""
		config.Get().Postgres.Get(&Voter, "select fullname from profiles where id = $1 limit 1", table.VoterID)
		tableMap["Voter"] = Voter
		
		message.Body = tableMap
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiVotesPost(httpRes http.ResponseWriter, httpReq *http.Request) {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Votes{}
		table.FillStruct(tableMap)

		if table.Title == "" {
			message.Message += "Title is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		if table.Workflow == "" {
			message.Message += "Status is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		if table.ProposalID == uint64(0) {
			message.Message += "Proposal is required \n"
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

		if table.CandidateID == uint64(0) {
			message.Message += "Candidate is required \n"
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

func apiVotesSearch(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureSearch(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Votes{}
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

			Candidate := ""
			config.Get().Postgres.Get(&Candidate, "select fullname from profiles where id = $1 limit 1", result.CandidateID)
			tableMap["Candidate"] = Candidate

			Voter := ""
			config.Get().Postgres.Get(&Voter, "select fullname from profiles where id = $1 limit 1", result.VoterID)
			tableMap["Voter"] = Voter

			searchList = append(searchList, tableMap)
		}
		message.Body = searchList

	}
	json.NewEncoder(httpRes).Encode(message)
}

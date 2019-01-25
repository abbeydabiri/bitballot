package api

import (
	"encoding/json"
	"net/http"

	"github.com/justinas/alice"

	"bitballot/database"
)

func apiHandlerProposals(middlewares alice.Chain, router *Router) {
	router.Get("/api/proposals", middlewares.ThenFunc(apiProposalsGet))
	router.Post("/api/proposals", middlewares.ThenFunc(apiProposalsPost))
	router.Post("/api/proposals/search", middlewares.ThenFunc(apiProposalsSearch))
}

func apiProposalsGet(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Proposals{}
		table.GetByID(table.ToMap(), formSearch)
		message.Body = table.ToMap()
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiProposalsPost(httpRes http.ResponseWriter, httpReq *http.Request) {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Proposals{}
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

func apiProposalsSearch(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureSearch(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Proposals{}
		if formSearch.Field == "" {
			formSearch.Field = "Title"
		}
		var searchList []interface{}
		searchResults := table.Search(table.ToMap(), formSearch)
		for _, result := range searchResults {
			searchList = append(searchList, result.ToMap())
		}
		message.Body = searchList

	}
	json.NewEncoder(httpRes).Encode(message)
}

package api

import (
	"encoding/json"
	"net/http"

	"github.com/justinas/alice"

	"bitballot/database"
	"bitballot/utils"
	"bitballot/config"
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

		tableMap := table.ToMap()

		if !table.OpenDate.IsZero() {
			tableMap["OpenDateDay"], tableMap["OpenDateTime"] = utils.DateTimeSplit(table.OpenDate)
		}

		if !table.EndDate.IsZero() {
			tableMap["EndDateDay"], tableMap["EndDateTime"] = utils.DateTimeSplit(table.EndDate)
		}

		message.Body = tableMap
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiProposalsPost(httpRes http.ResponseWriter, httpReq *http.Request) {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Proposals{}
		table.FillStruct(tableMap)

		if table.Workflow == "" {
			message.Message += "Status is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		if table.Title == "" {
			message.Message += "Title is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		if table.OpenDateDay == "" {
			message.Message += "Open Date is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}
		
		if table.OpenDateTime == "" {
			message.Message += "Open Time is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}
		table.OpenDate = utils.DateTimeMerge(table.OpenDateDay, table.OpenDateTime, config.Get().Timezone)
		
		if table.OpenDate.IsZero() {
			message.Message += "Open Date & Time is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		//---

		if table.EndDateDay == "" {
			message.Message += "End Date is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}
		
		if table.EndDateTime == "" {
			message.Message += "End Time is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}
		table.EndDate = utils.DateTimeMerge(table.EndDateDay, table.EndDateTime, config.Get().Timezone)
		
		if table.EndDate.IsZero()  {
			message.Message += "End Date & Time is required \n"
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

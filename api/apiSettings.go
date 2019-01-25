package api

import (
	"encoding/json"
	"net/http"

	"github.com/justinas/alice"

	"bitballot/config"
	"bitballot/database"
)

func apiHandlerSettings(middlewares alice.Chain, router *Router) {
	router.Get("/api/settings", middlewares.ThenFunc(apiSettingsGet))
	router.Post("/api/settings", middlewares.ThenFunc(apiSettingsPost))
	router.Post("/api/settings/search", middlewares.ThenFunc(apiSettingsSearch))
}

func apiSettingsGet(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Settings{}
		table.GetByID(table.ToMap(), formSearch)

		tableMap := table.ToMap()

		Owner := ""
		config.Get().Postgres.Get(&Owner, "select fullname from profiles where id = $1 limit 1", table.OwnerID)
		tableMap["Owner"] = Owner

		Partner := ""
		config.Get().Postgres.Get(&Partner, "select fullname from profiles where id = $1 limit 1", table.PartnerID)
		tableMap["Partner"] = Partner

		message.Body = tableMap
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiSettingsPost(httpRes http.ResponseWriter, httpReq *http.Request) {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Settings{}
		table.FillStruct(tableMap)

		if table.Code == "" {
			message.Message += "Code is required \n"
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

		if table.Workflow == "" {
			message.Message += "Workflowpo is required \n"
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

func apiSettingsSearch(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureSearch(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Settings{}
		if formSearch.Field == "" {
			formSearch.Field = "Title"
		}
		var searchList []interface{}
		searchResults := table.Search(table.ToMap(), formSearch)
		for _, result := range searchResults {
			tableMap := result.ToMap()

			Owner := ""
			config.Get().Postgres.Get(&Owner, "select fullname from profiles where id = $1 limit 1", result.OwnerID)
			tableMap["Owner"] = Owner

			Partner := ""
			config.Get().Postgres.Get(&Partner, "select fullname from profiles where id = $1 limit 1", result.PartnerID)
			tableMap["Partner"] = Partner

			searchList = append(searchList, tableMap)
		}
		message.Body = searchList

	}
	json.NewEncoder(httpRes).Encode(message)
}

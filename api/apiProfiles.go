package api

import (
	"encoding/json"
	"net/http"
	"fmt"

	"bitballot/config"
	"bitballot/database"
	"bitballot/utils"

	"github.com/justinas/alice"
)

func apiHandlerProfiles(middlewares alice.Chain, router *Router) {
	router.Get("/api/profiles", middlewares.ThenFunc(apiProfilesGet))
	router.Post("/api/profiles", middlewares.ThenFunc(apiProfilesPost))
	router.Post("/api/profiles/search", middlewares.ThenFunc(apiProfilesSearch))
	router.Get("/api/profiles/delete", middlewares.ThenFunc(apiProfilesDelete))
}

func apiProfilesGet(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiGenericGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Profiles{}
		table.GetByID(table.ToMap(), formSearch)

		tableMap := table.ToMap()
		message.Body = tableMap
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiProfilesPost(httpRes http.ResponseWriter, httpReq *http.Request) {
	tableMap, message := apiGenericPost(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Profiles{}
		table.FillStruct(tableMap)

		if table.Image = utils.SaveBase64Image(table.Image, ""); table.Image != "" {
			tableMap["Image"] = table.Image
		}

		if table.Title != "" {
			table.Fullname = table.Title
		}

		if table.Lastname != "" {
			table.Fullname = fmt.Sprintf("%s %s", table.Fullname, table.Lastname)
		}

		if table.Firstname != "" {
			table.Fullname = fmt.Sprintf("%s %s", table.Fullname, table.Firstname)
		}

		if table.Othername != "" {
			table.Fullname = fmt.Sprintf("%s %s", table.Fullname, table.Othername)
		}

		if table.Fullname == "" {
			message.Message += "Fullname is required \n"
			message.Code = http.StatusInternalServerError
			json.NewEncoder(httpRes).Encode(message)
			return
		}

		if table.Workflow == "" {
			message.Message += "Workflow is required \n"
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

func apiProfilesSearch(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiGenericSearch(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Profiles{}
		if formSearch.Field == "" {
			formSearch.Field = "Title"
		}
		message.Body = table.Search(table.ToMap(), formSearch)
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiProfilesDelete(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		if formSearch.ID > 0 {
			sqlParams := []interface{}{formSearch.ID}
			sqlQuery := "delete from profiles where id = $1 "
			_, err := config.Get().Postgres.Exec(sqlQuery, sqlParams...)
			message.Message = "Profile Deleted!!"
			if err != nil {
				message.Message = err.Error()
			}
		} else {
			if message.Message == "" {
				message.Message = "Unable to delete Profile!!"
			}
		}

	}
	json.NewEncoder(httpRes).Encode(message)
}

package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/justinas/alice"
	"golang.org/x/crypto/bcrypt"

	"bitballot/config"
	"bitballot/database"
	"bitballot/utils"
)

func apiHandlerUsers(middlewares alice.Chain, router *Router) {
	router.Get("/api/users", middlewares.ThenFunc(apiUsersGet))
	router.Post("/api/users", middlewares.ThenFunc(apiUsersPost))
	router.Post("/api/users/search", middlewares.ThenFunc(apiUsersSearch))
	router.Get("/api/users/delete", middlewares.ThenFunc(apiUsersDelete))
}

func apiUsersGet(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Users{}
		table.GetByID(table.ToMap(), formSearch)

		tableMap := table.ToMap()
		tableMap["Password"] = ""

		message.Body = tableMap
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiUsersPost(httpRes http.ResponseWriter, httpReq *http.Request) {
	tableMap, message := apiSecurePost(httpRes, httpReq)

	if message.Code != http.StatusOK {
		json.NewEncoder(httpRes).Encode(message)
		return
	}

	table := database.Users{}
	table.FillStruct(tableMap)

	if table.Username == "" {
		message.Message += "Username is required \n"
		message.Code = http.StatusInternalServerError
		json.NewEncoder(httpRes).Encode(message)
		return
	}

	if table.Email == "" {
		message.Message += "Email is required \n"
		message.Code = http.StatusInternalServerError
		json.NewEncoder(httpRes).Encode(message)
		return
	}

	if table.Image = utils.SaveBase64Image(table.Image, ""); table.Image != "" {
		tableMap["Image"] = table.Image
	}

	if tableMap["Password"] != nil && tableMap["Password"].(string) != "" {
		passwordHash, _ := bcrypt.GenerateFromPassword(
			[]byte(tableMap["Password"].(string)), bcrypt.DefaultCost)
		tableMap["Password"] = base64.StdEncoding.EncodeToString(passwordHash)
	}

	if table.ID == 0 {
		table.FillStruct(tableMap)
		table.Create(table.ToMap())
	} else {
		table.Update(tableMap)
	}
	message.Body = table.ID
	message.Message = RecordSaved
	json.NewEncoder(httpRes).Encode(message)
}

func apiUsersSearch(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureSearch(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Users{}
		if formSearch.Field == "" {
			formSearch.Field = "Username"
		}

		var searchList []interface{}
		searchResults := table.Search(table.ToMap(), formSearch)
		for _, result := range searchResults {
			tableMap := result.ToMap()
			searchList = append(searchList, tableMap)
		}
		message.Body = searchList

	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiUsersDelete(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		if formSearch.ID > 0 {
			sqlParams := []interface{}{formSearch.ID}
			sqlQuery := "delete from users where id = $1 "
			_, err := config.Get().Postgres.Exec(sqlQuery, sqlParams...)
			message.Message = "User Deleted!!"
			if err != nil {
				message.Message = err.Error()
			}
		} else {
			if message.Message == "" {
				message.Message = "Unable to delete User!!"
			}
		}

	}
	json.NewEncoder(httpRes).Encode(message)
}

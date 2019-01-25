package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/justinas/alice"

	"bitballot/config"
	"bitballot/database"
	"bitballot/utils"
)

func apiHandlerDocuments(middlewares alice.Chain, router *Router) {
	router.Get("/api/documents", middlewares.ThenFunc(apiDocumentsGet))
	router.Post("/api/documents", middlewares.ThenFunc(apiDocumentsPost))
	router.Post("/api/documents/search", middlewares.ThenFunc(apiDocumentsSearch))
	router.Get("/api/documents/delete", middlewares.ThenFunc(apiDocumentsDelete))
}

type documentItem struct {
	ID          uint64
	File, Title string
}

func apiDocumentsGet(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Documents{}
		table.GetByID(table.ToMap(), formSearch)

		tableMap := table.ToMap()

		Owner := ""
		config.Get().Postgres.Get(&Owner, "select fullname from profiles where id = $1 limit 1", table.OwnerID)
		tableMap["Owner"] = Owner

		message.Body = tableMap
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiDocumentsPost(httpRes http.ResponseWriter, httpReq *http.Request) {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Documents{}
		table.FillStruct(tableMap)

		if table.Title == "" {
			message.Message += "Title is required \n"
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

		table.Filename = fmt.Sprintf("%s-%s", utils.RandomString(3), table.Filename)
		if table.Filepath = utils.SaveBase64File(table.Filepath, "files", table.Filename); table.Filepath != "" {
			tableMap["Filepath"] = table.Filepath
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

func apiDocumentsSearch(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureSearch(httpRes, httpReq)
	if message.Code == http.StatusOK {
		table := database.Documents{}
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

			searchList = append(searchList, tableMap)
		}
		message.Body = searchList

	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiDocumentsTableList(TableRef, Doctype string, TableRefID uint64) (documentItemList map[uint64]documentItem) {
	var documentList []documentItem
	documentItemList = make(map[uint64]documentItem)
	sqlQuery := "select id, title, filepath as file from documents where tableref = $1 and " +
		"doctype = $2 and tablerefid = $3 and workflow = 'enabled' order by title"
	if err := config.Get().Postgres.Select(&documentList, sqlQuery, TableRef, Doctype, TableRefID); err != nil {
		log.Println(err.Error())
		return
	}

	for _, item := range documentList {
		documentItemList[item.ID] = item
	}
	return
}

func apiDocumentList(TableRef, Doctype string, TableRefID uint64) (workflowList map[uint64]string) {
	type workflowListStruct struct {
		ID       uint64
		Workflow string
	}
	var documentList []workflowListStruct
	workflowList = make(map[uint64]string)
	sqlQuery := "select id, workflow from documents where tableref = $1 and " +
		"doctype = $2 and tablerefid = $3 and workflow = 'enabled' order by title"
	if err := config.Get().Postgres.Select(&documentList, sqlQuery, TableRef, Doctype, TableRefID); err != nil {
		log.Println(err.Error())
		return
	}

	for _, item := range documentList {
		workflowList[item.ID] = item.Workflow
	}
	return
}

func apiDocumentListSave(TableRef, Doctype string, TableRefID uint64, documentsList []database.Documents) (statusMessage string) {
	curdocumentsList := []database.Documents{}
	sqlQuery := "select id, ownerid, filename, filetype, filepath, filesize, position from documents " +
		"where tableref = $1 and doctype = $2 and tablerefid = $3 and workflow = 'enabled' order by title"
	if err := config.Get().Postgres.Select(&curdocumentsList, sqlQuery, TableRef, Doctype, TableRefID); err != nil {
		log.Println(err.Error())
		return
	}

	documentsListSave := []database.Documents{}
	documentsListSearch := make(map[string]database.Documents)

	for _, result := range curdocumentsList {
		documentsListSearch[fmt.Sprintf("%v", result.ID)] = result
	}

	table := database.Documents{}
	for _, documentItem := range documentsList {

		if documentsListSearch[fmt.Sprintf("%v", documentItem.ID)].ID > 0 {
			table = documentsListSearch[fmt.Sprintf("%v", documentItem.ID)]
			table.Filename = documentItem.Filename
			table.Filetype = documentItem.Filetype
			table.Filepath = documentItem.Filepath
			table.Filesize = documentItem.Filesize
			table.Position = documentItem.Position
		} else {
			table = database.Documents{}
			table.ID = 0
			table.Workflow = "enabled"
		}

		table.TableRefID = TableRefID
		table.TableRef = TableRef
		table.Doctype = Doctype

		documentsListSave = append(documentsListSave, table)
		delete(documentsListSearch, fmt.Sprintf("%v", table.ID))
	}

	//UPDATE Documents HERE
	//First Delete Old Documents
	var documentINlist []uint64
	for _, document := range documentsListSearch {
		documentINlist = append(documentINlist, document.ID)
	}
	config.Get().Postgres.Exec("delete from documents where id in ($1)", documentINlist)
	//First Delete Old Documents

	//Now Create New Documents
	for _, document := range documentsListSave {
		if document.ID == 0 {
			document.Create(document.ToMap())
		} else {
			document.Update(document.ToMap())
		}
	}
	//Now Create New Documents
	//UPDATE Documents HERE

	return
}

func apiDocumentsDelete(httpRes http.ResponseWriter, httpReq *http.Request) {
	formSearch, message := apiSecureGet(httpRes, httpReq)
	if message.Code == http.StatusOK {
		if formSearch.ID > 0 {
			sqlParams := []interface{}{formSearch.ID}
			sqlQuery := "delete from documents where id = $1 "
			_, err := config.Get().Postgres.Exec(sqlQuery, sqlParams...)
			message.Message = "Document Deleted!!"
			if err != nil {
				message.Message = err.Error()
			}
		} else {
			if message.Message == "" {
				message.Message = "Unable to delete Document!!"
			}
		}

	}
	json.NewEncoder(httpRes).Encode(message)
}

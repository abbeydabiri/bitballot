package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/rs/cors"
)

func apiInit() {
	router := httprouter.New()
	middlewares := alice.New()

	//handle existing api endpoints
	router.GET("/api/get", apiGet)
	router.POST("/api/post", apiPost)

	// load spa app by default
	router.NotFound = middlewares.ThenFunc(
		func(httpRes http.ResponseWriter, httpReq *http.Request) {
			spaFS(httpRes, httpReq)
		},
	)

	mainHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*", "localhost"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Auth-Token", "*"},
	}).Handler(router)

	println("serving @ " + Config.Address)
	http.ListenAndServe(Config.Address, mainHandler)
}

func apiGet(httpRes http.ResponseWriter, httpReq *http.Request, _ httprouter.Params) {
	httpRes.Header().Set("Content-Type", "application/json")

	strLen := 8
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	byteResult := make([]byte, strLen)
	for i := 0; i < strLen; i++ {
		byteResult[i] = chars[rand.Intn(len(chars))]
	}

	shaCrypt := sha1.New()
	shaCrypt.Write([]byte(byteResult))

	uuidString := hex.EncodeToString(shaCrypt.Sum(nil))
	uuidString = fmt.Sprintf("%s-%s-%s-%s-%s", uuidString[:8], uuidString[9:12], uuidString[13:16], uuidString[17:20], uuidString[21:])
	jsonResponse := map[string]string{"uuid": uuidString}

	json.NewEncoder(httpRes).Encode(jsonResponse)

}

type errorResponseJson struct {
	Error string
}

func apiPost(httpRes http.ResponseWriter, httpReq *http.Request, _ httprouter.Params) {
	httpRes.Header().Set("Content-Type", "application/json")

	errorResponse := &errorResponseJson{}

	var requestMap struct {
		UUID, File, Filename string
	}

	if err := json.NewDecoder(httpReq.Body).Decode(&requestMap); err != nil {
		errorResponse.Error += "Error Decoding Form Values: " + err.Error()
		json.NewEncoder(httpRes).Encode(errorResponse)
		return
	}

	if requestMap.Filename == "" {
		errorResponse.Error = "Filename is missing! "
		json.NewEncoder(httpRes).Encode(errorResponse)
		return
	}

	if requestMap.UUID == "" {
		errorResponse.Error = "Customer UUID is missing! "
		json.NewEncoder(httpRes).Encode(errorResponse)
		return
	}

	base64Bytes, err := base64.StdEncoding.DecodeString(strings.Split(requestMap.File, "base64,")[1])

	if err != nil {
		errorResponse.Error = "Error Decoding CSV file " + err.Error()
		json.NewEncoder(httpRes).Encode(errorResponse)
		return
	}
	if base64Bytes == nil {
		errorResponse.Error = "CSV file is empty"
		json.NewEncoder(httpRes).Encode(errorResponse)
		return
	}

	pipeCSV := string(base64Bytes)
	pipeCSV = strings.Replace(pipeCSV, "\r", "", -1)
	pipeCSV = strings.Replace(pipeCSV, "\n\n", "\n", -1)
	sliceRow := strings.Split(pipeCSV, "\n")

	//@TODO loop through csv content and generate single api call
	var riskRequestSlice []interface{}
	for _, stringCols := range sliceRow {
		riskRequestItem := make(map[string]string)
		if len(strings.TrimSpace(stringCols)) == 0 {
			continue
		}

		sliceCols := strings.Split(strings.TrimSpace(stringCols), ",")
		for index, value := range sliceCols {
			value = strings.TrimSpace(value)
			switch index {
			case 0:
				riskRequestItem["asset"] = value
			case 1:
				riskRequestItem["address"] = value
			}
		}

		riskRequestSlice = append(riskRequestSlice, riskRequestItem)
	}

	sURL := fmt.Sprintf("/users/%s/withdrawaladdresses", requestMap.UUID)
	jsonRequest, _ := json.Marshal(riskRequestSlice)

	log.Printf("%s\n", jsonRequest)

	errorMessage, jsonResponse := curl("POST", sURL, jsonRequest)

	if errorMessage != "" {
		errorResponse.Error = errorMessage
		json.NewEncoder(httpRes).Encode(errorResponse)
		return
	}

	//@TODO retrieve the respone and send to user interface
	type riskResponse struct {
		Asset, Address,
		Rating string
	}
	jsonResponseSlice := []riskResponse{}
	json.Unmarshal([]byte(string(jsonResponse)), &jsonResponseSlice)
	json.NewEncoder(httpRes).Encode(jsonResponseSlice)

}

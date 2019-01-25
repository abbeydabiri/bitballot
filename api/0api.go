package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bitballot/config"
	"bitballot/database"
	"bitballot/utils"

	jwt "github.com/dgrijalva/jwt-go"
)

func apiGeneric(httpRes http.ResponseWriter, httpReq *http.Request) (message *Message) {
	httpRes.Header().Set("Content-Type", "application/json")
	message = new(Message)
	message.Code = http.StatusInternalServerError
	message.Code = http.StatusOK

	claims := utils.VerifyJWT(httpRes, httpReq)
	if claims == nil {
		cookieExpires := time.Now().Add(time.Minute * 15) // set the expire time
		jwtClaims := map[string]interface{}{
			"exp": cookieExpires.Unix(),
		}

		if jwtToken, err := utils.GenerateJWT(jwtClaims); err == nil {
			cookieMonster := &http.Cookie{
				Name: config.Get().COOKIE, Value: jwtToken, Expires: cookieExpires, Path: "/",
			}
			http.SetCookie(httpRes, cookieMonster)
			httpReq.AddCookie(cookieMonster)
		}
	}
	return
}

func apiGenericSearch(httpRes http.ResponseWriter, httpReq *http.Request) (formSearch *database.SearchParams, message *Message) {
	httpRes.Header().Set("Content-Type", "application/json")
	message = new(Message)
	formSearch = new(database.SearchParams)
	message.Code = http.StatusInternalServerError

	claims := utils.VerifyJWT(httpRes, httpReq)
	if claims == nil {
		log.Println("ERROR: Claim is nil")
		message.Body = map[string]string{"Redirect": "/"}
		return
	}

	err := json.NewDecoder(httpReq.Body).Decode(formSearch)
	if err != nil {
		message.Message = "Error Decoding Form Values: " + err.Error()
		log.Println("ERROR: " + err.Error())
		return
	}

	formSearch.Text = strings.TrimSpace(formSearch.Text)
	message.Code = http.StatusOK
	return
}

func apiGenericGet(httpRes http.ResponseWriter, httpReq *http.Request) (formSearch *database.SearchParams, message *Message) {
	message = apiGeneric(httpRes, httpReq)
	formSearch = new(database.SearchParams)

	claims := utils.VerifyJWT(httpRes, httpReq)
	if claims == nil {
		message.Body = map[string]string{"Redirect": "/"}
		return
	}

	sID := strings.TrimSpace(httpReq.FormValue("id"))
	if sID == "" {
		message.Code = http.StatusInternalServerError
		message.Message = "ID is required"
		return
	}

	formSearch.ID, _ = strconv.ParseUint(sID, 0, 64)
	message.Code = http.StatusOK
	return
}

func apiGenericPost(httpRes http.ResponseWriter, httpReq *http.Request) (tableMap map[string]interface{}, message *Message) {
	message = apiGeneric(httpRes, httpReq)
	tableMap = make(map[string]interface{})

	claims := utils.VerifyJWT(httpRes, httpReq)
	if claims == nil {
		message.Body = map[string]string{"Redirect": "/"}
		return
	}

	jsonBody, err := ioutil.ReadAll(httpReq.Body)
	defer httpReq.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return
	}

	if err = json.Unmarshal(jsonBody, &tableMap); err != nil {
		log.Println(err.Error())
		return
	}

	message.Code = http.StatusOK
	return
}

func apiSecure(httpRes http.ResponseWriter, httpReq *http.Request) (message *Message) {
	httpRes.Header().Set("Content-Type", "application/json")
	message = new(Message)
	message.Code = http.StatusInternalServerError
	claims := utils.VerifyJWT(httpRes, httpReq)
	if claims == nil {
		message.Body = map[string]string{"Redirect": "/"}
		return
	}
	if apiBlock("admin", claims) && apiBlock("support", claims) && apiBlock("client", claims) {
		apiBlockResponse(httpRes)
		return
	}
	message.Code = http.StatusOK
	return
}

func apiSecureSearch(httpRes http.ResponseWriter, httpReq *http.Request) (formSearch *database.SearchParams, message *Message) {
	httpRes.Header().Set("Content-Type", "application/json")
	message = new(Message)
	formSearch = new(database.SearchParams)
	message.Code = http.StatusInternalServerError

	claims := utils.VerifyJWT(httpRes, httpReq)
	if claims == nil {
		message.Body = map[string]string{"Redirect": "/"}
		return
	}

	if apiBlock("admin", claims) && apiBlock("support", claims) && apiBlock("client", claims) {
		apiBlockResponse(httpRes)
		return
	}

	err := json.NewDecoder(httpReq.Body).Decode(formSearch)
	if err != nil {
		message.Message = "Error Decoding Form Values: " + err.Error()
		log.Println("ERROR: " + err.Error())
		return
	}

	formSearch.Text = strings.TrimSpace(formSearch.Text)
	message.Code = http.StatusOK
	return
}

func apiSecurePost(httpRes http.ResponseWriter, httpReq *http.Request) (tableMap map[string]interface{}, message *Message) {
	httpRes.Header().Set("Content-Type", "application/json")
	message = new(Message)
	message.Code = http.StatusInternalServerError

	tableMap = make(map[string]interface{})
	claims := utils.VerifyJWT(httpRes, httpReq)
	if claims == nil {
		message.Body = map[string]string{"Redirect": "/"}
		return
	}

	if apiBlock("admin", claims) && apiBlock("support", claims) && apiBlock("client", claims) {
		apiBlockResponse(httpRes)
		return
	}

	jsonBody, err := ioutil.ReadAll(httpReq.Body)
	defer httpReq.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return
	}

	if err = json.Unmarshal(jsonBody, &tableMap); err != nil {
		log.Println(err.Error())
		return
	}

	if claims["ID"] != nil {
		tableMap["Updatedby"] = uint64(claims["ID"].(float64))
	}

	message.Code = http.StatusOK
	return
}

func apiSecureGet(httpRes http.ResponseWriter, httpReq *http.Request) (formSearch *database.SearchParams, message *Message) {
	httpRes.Header().Set("Content-Type", "application/json")
	message = new(Message)
	formSearch = new(database.SearchParams)
	message.Code = http.StatusInternalServerError

	claims := utils.VerifyJWT(httpRes, httpReq)
	if claims == nil {
		message.Body = map[string]string{"Redirect": "/"}
		return
	}

	if apiBlock("admin", claims) && apiBlock("support", claims) && apiBlock("client", claims) {
		apiBlockResponse(httpRes)
		return
	}

	sID := strings.TrimSpace(httpReq.FormValue("id"))
	if sID == "" {
		message.Code = http.StatusInternalServerError
		message.Message = "ID is required"
		return
	}

	formSearch.ID, _ = strconv.ParseUint(sID, 0, 64)
	message.Code = http.StatusOK
	return
}

func apiBlockResponse(httpRes http.ResponseWriter) {
	json.NewEncoder(httpRes).Encode(Message{
		Code:    http.StatusInternalServerError,
		Body:    nil,
		Message: "Permission Required",
	})
	return
}

func apiBlock(role string, claims jwt.MapClaims) (block bool) {
	block = true
	if claims != nil {
		switch role {
		case "admin":
			if claims["IsAdmin"] != nil && claims["IsAdmin"].(bool) {
				block = false
			}
		case "support":
			if claims["IsSupport"] != nil && claims["IsSupport"].(bool) {
				block = false
			}
		case "client":
			if claims["IsCustomer"] != nil && claims["IsCustomer"].(bool) {
				block = false
			}
		}
	}
	return
}

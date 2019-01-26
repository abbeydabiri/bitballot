package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"


	// "bytes"
	// "html/template"
	// "log"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"

	"bitballot/config"
	"bitballot/database"
	"bitballot/utils"
	
)

const (
	IsRequired = " is Required \n" //isRequired is required

	RecordSaved = "Record Saved" //recordSaved is saved

	TitleRequired    = "Title is Required \n"
	WorkflowRequired = "Workflow is Required \n"
)

//Message ...
type Message struct {
	Code           int
	Message, Error string
	Body           interface{}
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("02/01/2006 03:04:05 PM"))
	return []byte(stamp), nil
}

func verifyID(httpRes http.ResponseWriter, httpReq *http.Request, claims jwt.MapClaims) {
	if claims == nil {
		http.Redirect(httpRes, httpReq, "/", http.StatusTemporaryRedirect)
		return
	}

	if claims["ID"] == nil {
		http.Redirect(httpRes, httpReq, "/", http.StatusTemporaryRedirect)
		return
	}
}

func apiHandler(middlewares alice.Chain, router *Router) {

	router.POST("/db/init/*table", func(httpRes http.ResponseWriter, httpReq *http.Request, httpParams httprouter.Params) {
		claims := utils.VerifyJWT(httpRes, httpReq)
		if claims == nil {
			http.Redirect(httpRes, httpReq, "/404.html", http.StatusTemporaryRedirect)
			return
		}

		if claims["IsAdmin"].(bool) && claims["Username"].(string) == "root" {
			errMessages := database.Init(httpParams.ByName("table"))

			message := new(Message)
			message.Code = http.StatusOK
			message.Message = strings.Join(errMessages, "\n")
			json.NewEncoder(httpRes).Encode(message)
			return
		} else {
			http.Redirect(httpRes, httpReq, "/404.html", http.StatusTemporaryRedirect)
		}

	})

	//Authenticated Pages --> Below
	router.GET("/login", func(httpRes http.ResponseWriter, httpReq *http.Request, httpParams httprouter.Params) {
		if claims := utils.VerifyJWT(httpRes, httpReq); claims != nil {
			if uint64(claims["ID"].(float64)) > 0 {
				switch {
				case claims["IsAdmin"] != nil && claims["IsAdmin"].(bool):
					http.Redirect(httpRes, httpReq, "/admin", http.StatusTemporaryRedirect)
					break

				case claims["IsSupport"] != nil && claims["IsSupport"].(bool):
					http.Redirect(httpRes, httpReq, "/support", http.StatusTemporaryRedirect)
					break

				default:
					http.Redirect(httpRes, httpReq, "/dashboard", http.StatusTemporaryRedirect)
					break
				}
			}
		}
		fileServe(httpRes, httpReq)
	})

	router.GET("/admin/*page", func(httpRes http.ResponseWriter, httpReq *http.Request, httpParams httprouter.Params) {
		claims := utils.VerifyJWT(httpRes, httpReq)
		verifyID(httpRes, httpReq, claims)
		// if claims["IsAdmin"] == nil || !claims["IsAdmin"].(bool) {
		// 	http.Redirect(httpRes, httpReq, "/", http.StatusTemporaryRedirect)
		// 	return
		// }
		fileServe(httpRes, httpReq)
	})

	//Authenticated Pages --> Below
	apiHandlerAuth(middlewares, router)

	apiHandlerDashboard(middlewares, router)

	apiHandlerCandidates(middlewares, router)
	apiHandlerDocuments(middlewares, router)
	
	apiHandlerPositions(middlewares, router)
	apiHandlerProfiles(middlewares, router)
	apiHandlerProposals(middlewares, router)
	
	apiHandlerSettings(middlewares, router)
	apiHandlerUsers(middlewares, router)
	apiHandlerVoters(middlewares, router)
	apiHandlerVotes(middlewares, router)
	
	apiHandlerEthreum(middlewares, router)
}

func cut(name string) string {
	name = strings.TrimSuffix(name, "/")
	dir, _ := path.Split(name)
	return dir
}

func fileServe(httpRes http.ResponseWriter, httpReq *http.Request) {

	urlPath := strings.Replace(httpReq.URL.Path, "//", "/", -1)
	if strings.HasSuffix(urlPath, "/") {
		urlPath = path.Join(urlPath, "index.html")
	}

	var err error
	var dataBytes []byte

	if dataBytes, err = config.Asset("frontend" + urlPath); err != nil {
		for urlPath != "/" {
			urlPath = cut(urlPath)
			newPath := path.Join(urlPath, "index.html")
			if dataBytes, err = config.Asset("frontend" + newPath); err == nil {
				break
			}
		}
	}

	httpRes.Header().Set("Cache-Control", "max-age=0, must-revalidate")
	httpRes.Header().Set("Pragma", "no-cache")
	httpRes.Header().Set("Expires", "0")

	httpRes.Header().Add("Content-Type", config.ContentType(urlPath))
	if !strings.Contains(httpReq.Header.Get("Accept-Encoding"), "gzip") {
		httpRes.Write(dataBytes)
		return
	}
	gzipWrite(dataBytes, httpRes)
}

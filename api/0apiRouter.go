package api

import (
	"compress/gzip"
	"context"
	"io"
	"strings"
	"time"

	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/rs/cors"

	"bitballot/config"
	"bitballot/database"
)

//Router ...
type Router struct { // Router struct would carry the httprouter instance,
	*httprouter.Router //so its methods could be overwritten and replaced with methds with wraphandler
}

//Get ...
func (router *Router) Get(path string, handler http.Handler) {
	router.GET(path, wrapHandler(handler)) // Get is an endpoint to only accept requests of method GET
}

//Post is an endpoint to only accept requests of method POST
func (router *Router) Post(path string, handler http.Handler) {
	router.POST(path, wrapHandler(handler))
}

//Put is an endpoint to only accept requests of method PUT
func (router *Router) Put(path string, handler http.Handler) {
	router.PUT(path, wrapHandler(handler))
}

//Delete is an endpoint to only accept requests of method DELETE
func (router *Router) Delete(path string, handler http.Handler) {
	router.DELETE(path, wrapHandler(handler))
}

//NewRouter is a wrapper that makes the httprouter struct a child of the router struct
func NewRouter() *Router {
	return &Router{httprouter.New()}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipWrite(dataBytes []byte, httpRes http.ResponseWriter) {
	//httpRes.Header().Set("Transfer-Encoding", "gzip")
	httpRes.Header().Set("Content-Encoding", "gzip")
	gzipHandler := gzip.NewWriter(httpRes)
	defer gzipHandler.Close()
	httpResGzip := gzipResponseWriter{Writer: gzipHandler, ResponseWriter: httpRes}
	httpResGzip.Write(dataBytes)
}

func wrapHandler(httpHandler http.Handler) httprouter.Handle {
	return func(httpRes http.ResponseWriter, httpReq *http.Request, httpParams httprouter.Params) {
		ctx := context.WithValue(httpReq.Context(), "params", httpParams)
		httpReq = httpReq.WithContext(ctx)

		if !strings.Contains(httpReq.Header.Get("Accept-Encoding"), "gzip") {
			httpHandler.ServeHTTP(httpRes, httpReq)
			return
		}

		httpRes.Header().Set("Content-Encoding", "gzip")
		gzipHandler := gzip.NewWriter(httpRes)
		defer gzipHandler.Close()
		httpResGzip := gzipResponseWriter{Writer: gzipHandler, ResponseWriter: httpRes}
		httpHandler.ServeHTTP(httpResGzip, httpReq)
	}
}

//StartRouter ...
func StartRouter() {

	totalUsers := 0
	if config.Get().Postgres.Get(&totalUsers, "select count(id) from users"); totalUsers == 0 {
		database.Init("/all")
	}

	middlewares := alice.New()
	router := NewRouter()

	apiHandler(middlewares, router)

	router.NotFound = middlewares.ThenFunc(
		func(httpRes http.ResponseWriter, httpReq *http.Request) {
			frontend := strings.Split(httpReq.URL.Path[1:], "/")
			switch frontend[0] {
			case "logout":
				cookieMonster := &http.Cookie{
					Name: config.Get().COOKIE, Value: "deleted", Path: "/",
					Expires: time.Now().Add(-(time.Hour * 24 * 30 * 12)), // set the expire time
				}
				http.SetCookie(httpRes, cookieMonster)
				httpReq.AddCookie(cookieMonster)
				http.Redirect(httpRes, httpReq, "/", http.StatusTemporaryRedirect)

			default:
				fileServe(httpRes, httpReq)
			}
		})

	mainHandler := cors.New(cors.Options{
		// AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Auth-Token", "*"},
	}).Handler(router)

	sMessage := "serving @ " + config.Get().Address
	println(sMessage)
	log.Println(sMessage)
	log.Fatal(http.ListenAndServe(config.Get().Address, mainHandler))
}

package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/justinas/alice"
)

//apiHandlerWeb3 ...
func apiHandlerWeb3(middlewares alice.Chain, router *Router) {
	router.Post("/web3", middlewares.ThenFunc(apiWeb3))
}

func apiWeb3(httpRes http.ResponseWriter, httpReq *http.Request) {
	body, err := ioutil.ReadAll(httpReq.Body)
	if err != nil {
		http.Error(httpRes, err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Printf("\n %s", body)

	// you can reassign the body if you need to parse it as multipart
	httpReq.Body = ioutil.NopCloser(bytes.NewReader(body))

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s", "https", "rinkeby.infura.io", "/wvxLGQSZBjP3Ak7iqt8J")

	proxyReq, err := http.NewRequest(httpReq.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = httpReq.Header
	proxyReq.Header = make(http.Header)
	for h, val := range httpReq.Header {
		proxyReq.Header[h] = val
	}

	tlsConf := &tls.Config{}
	tr := &http.Transport{TLSClientConfig: tlsConf}
	httpClient := &http.Client{Transport: tr}

	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		http.Error(httpRes, err.Error(), http.StatusBadGateway)
		fmt.Printf("%v \n", err.Error())
		return
	}

	defer resp.Body.Close()
	bodyResp, errResp := ioutil.ReadAll(resp.Body)

	if errResp != nil {
		fmt.Printf("\n errResp: %s \n", errResp.Error())
		return
	}
	// fmt.Printf("\n bodyResp: %s \n", bodyResp)

	httpRes.Header().Set("Content-Type", "application/json")
	httpRes.Write(bodyResp)

}

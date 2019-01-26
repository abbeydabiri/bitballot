package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"

	"github.com/justinas/alice"
	
	"bitballot/blockchain"
)

//apiHandlerEthreum ...
func apiHandlerEthreum(middlewares alice.Chain, router *Router) {
	//Connect to ETHEREUM BLOCKCHAIN
	go blockchain.EthClientDial()

	// router.Post("/web3", middlewares.ThenFunc(apiEthereum))
	router.Get("/api/eth/balance", middlewares.ThenFunc(apiEthereumBalance))
	router.Post("/api/eth/proposal", middlewares.ThenFunc(apiEthereumProposal))
	router.Post("/api/eth/position", middlewares.ThenFunc(apiEthereumPosition))
	router.Post("/api/eth/candidate", middlewares.ThenFunc(apiEthereumCandidate))
	router.Post("/api/eth/voters", middlewares.ThenFunc(apiEthereumVoters))
}

func apiEthereumProposal(httpRes http.ResponseWriter, httpReq *http.Request)  {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {

		table := database.Proposals{}
		table.GetByID(table.ToMap(), tableMap["id"])

		message.Body = nil
		message.Message = Sent to Blockchain
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiEthereumPosition(httpRes http.ResponseWriter, httpReq *http.Request)  {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {

		table := database.Proposals{}
		table.GetByID(table.ToMap(), tableMap["id"])

		message.Body = nil
		message.Message = Sent to Blockchain
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiEthereumCandidate(httpRes http.ResponseWriter, httpReq *http.Request)  {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {

		table := database.Proposals{}
		table.GetByID(table.ToMap(), tableMap["id"])

		message.Body = nil
		message.Message = Sent to Blockchain
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiEthereumVoters(httpRes http.ResponseWriter, httpReq *http.Request)  {
	tableMap, message := apiSecurePost(httpRes, httpReq)
	if message.Code == http.StatusOK {

		table := database.Proposals{}
		table.GetByID(table.ToMap(), tableMap["id"])

		message.Body = nil
		message.Message = Sent to Blockchain
	}
	json.NewEncoder(httpRes).Encode(message)
}

func apiEthereumBalance(httpRes http.ResponseWriter, httpReq *http.Request) {
	message := apiSecure(httpRes, httpReq)
	if message.Code == http.StatusOK {
		
		_, fromAddress := blockchain.EthGenerateKey(blockchain.ETHAddress)
		balance, _ := blockchain.ETHAccountBalFloat(fromAddress.Hex(), 0)
	
		tableMap := map[string]interface{}{
			"balance": balance,
			"address": fromAddress.Hex(),
		}
		message.Body = tableMap
	}
	json.NewEncoder(httpRes).Encode(message)	
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



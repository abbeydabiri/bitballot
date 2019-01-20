package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

func cut(name string) string {
	name = strings.TrimSuffix(name, "/")
	dir, _ := path.Split(name)
	return dir
}

//Single Page Application filesystem
func spaFS(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRoot := http.Dir(Config.Path)
	httpRootFS := http.FileServer(httpRoot)

	upath := httpReq.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}

	f, err := httpRoot.Open(upath)
	if err == nil {
		f.Close()
	}

	// search index.html in ancestor directories and return it if exists.
	if err != nil && os.IsNotExist(err) {
		for upath != "/" {
			upath = cut(upath)
			f, err := httpRoot.Open(path.Join(upath, "index.html"))
			switch {
			case err == nil:
				f.Close()
				fallthrough
			case !os.IsNotExist(err):
				httpReq.URL.Path = upath
				httpRootFS.ServeHTTP(httpRes, httpReq)
				return
			}
		}
	}
	httpRootFS.ServeHTTP(httpRes, httpReq)
}

func curl(sMethod, sUrl string, byteBody []byte) (string, []byte) {

	sUrl = Config.URL + sUrl
	mapAuthHeader := map[string]string{"Token": Config.Key, "Content-Type": "application/json"}
	if sUrl == "" {
		return "Either sUrl or sMethod is missing", nil
	}

	if sMethod == "" {
		return "Either sUrl or sMethod is missing", nil
	}

	httpReq, _ := http.NewRequest(sMethod, sUrl, bytes.NewBuffer(byteBody))
	httpReq.Header.Add("Content-Length", strconv.Itoa(len(byteBody)))

	for sKey, sValue := range mapAuthHeader {
		httpReq.Header.Add(sKey, sValue)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "curl error: " + err.Error(), nil
	}

	resBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return "", resBody
}

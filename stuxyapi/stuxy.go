package stuxy

import (
	"appengine"
	"appengine/datastore"
	//	"appengine/user"
	//	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {

	// Stub操作
	http.HandleFunc("/env/mode", PutModeOfStubPage)
	http.HandleFunc("/env/post", PostStubPage)
	http.HandleFunc("/env/list", getStubList)
	// GroupConfig操作
	http.HandleFunc("/env/config", handleConfig)
	http.HandleFunc("/env/config/list", handleConfigList)

	// Postしてみる機能
	http.HandleFunc("/env/fetch/postform", FetchURLHandle)

	// スタブ
	http.HandleFunc("/", fetchStubPage)
}

/** スタブデータを登録する */
func PostStubPage(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	switch r.Method {
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Connection", "close")
		return

	case "GET":
		fallthrough
	case "POST":

		title := r.FormValue("title")
		path := r.FormValue("path")
		data := r.FormValue("data")
		mode := r.FormValue("mode")
		group := r.FormValue("group")
		contenttype := r.FormValue("contenttype")

		stub := new(StubPageModel)
		stub.ContentType = contenttype
		stub.Path = path
		stub.Data = data
		stub.Title = title
		stub.Mode = mode
		stub.Group = group

		if err := PutToDataStore(c, stub); err != nil {
			// putエラー
			http.Error(w, "could not put to datastore", http.StatusInternalServerError)
			return
		}

	case "DELETE":

		path := r.FormValue("path")
		stub, err := GetStubPage(c, path)
		if stub != nil {
			if err2 := datastore.Delete(c, stub.Key(c)); err2 != nil {
				c.Errorf("can't delete:%s", err2)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

/** スタブデータを返す */
func PutModeOfStubPage(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	switch r.Method {
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Connection", "close")
		return

	case "GET":
		fallthrough
	case "POST":
		r.ParseForm()
		w.Header().Set("Access-Control-Allow-Origin", "*")

		path := r.FormValue("path")
		mode, modeok := pullFormValue(r, "mode")
		wait, waitok := pullFormValue(r, "wait")

		stub, err := GetStubPage(c, path)
		if stub != nil {

			if modeok {
				stub.Mode = mode
			}
			if waitok {
				stub.Wait, _ = strconv.Atoi(wait)
			}

			if err := PutToDataStore(c, stub); err != nil {
				// putエラー
				http.Error(w, "could not put to datastore", http.StatusInternalServerError)
				return
			}

		} else if err == nil {
			// 指定されたStubがないって話
			http.Error(w, "could not put to datastore", http.StatusBadRequest)
		} else {
			http.Error(w, "could not put to datastore", http.StatusInternalServerError)
		}
	}
}

func pullFormValue(r *http.Request, key string) (string, bool) {

	a, ok := r.Form[key]
	if ok == false {
		return "", false
	}
	if len(a) == 0 {
		return "", false
	}

	return a[0], true
}

//------------------------------------------

func fetchStubPage(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	switch r.Method {

	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Connection", "close")
		return
	}

	// RequestされたURLを取得する
	path := r.URL.Path

	// StubPageを取得する
	stub, err := GetStubPage(c, path)
	if err != nil {
		http.Error(w, "could not query from datastore", http.StatusInternalServerError)
		return
	}

	//----------
	// Not Found
	if stub == nil {
		// Not FoundならUnknown GroupでProxyの可能性
		config, _ := LoadGroupConfig(c, "unknown")
		if config != nil && len(config.ProxyURL) > 0 {
			posturl := config.ProxyURL + path
			ProxyURL2(c, w, posturl, r)
		} else {
			http.Error(w, "404", http.StatusNotFound)
		}
		return
	}

	//----------
	// Wait
	if stub.Wait > 0 {
		time.Sleep(time.Duration(stub.Wait) * 1000 * 1000 * 1000)
	}

	//----------
	// Modeのチェック
	switch {
	case strings.Index(stub.Mode, "MD404") >= 0:
		http.Error(w, "404", http.StatusNotFound)
		return
	case strings.Index(stub.Mode, "MDPROXY") >= 0:
		// Proxy先を取得
		config, _ := LoadGroupConfig(c, stub.Group)
		if config != nil && len(config.ProxyURL) > 0 {
			posturl := config.ProxyURL + path
			ProxyURL2(c, w, posturl, r)
		} else {
			http.Error(w, "404", http.StatusNotFound)
		}
		return
	default:
		// データを返却
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", stub.ContentType)

		fmt.Fprint(w, stub.Data)
	}
}

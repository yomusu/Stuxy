package stuxy

import (
	"appengine"
	//	"encoding/json"
	"appengine/urlfetch"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

/** 渡された引数に基づきPostリクエストを行う */
func FetchURLHandle(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	r.ParseForm()

	// 引数解析
	posturl := r.FormValue("url")
	data, perr := url.ParseQuery(r.FormValue("data"))
	if perr != nil || len(posturl) == 0 {
		// 不正な引数
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, perr.Error(), http.StatusBadRequest)
		return
	}

	ProxyURL(c, w, posturl, data)
}

func ProxyURL(c appengine.Context, w http.ResponseWriter, posturl string, data url.Values) {

	c.Infof("Proxy URL:   " + posturl)
	c.Infof("Proxy PARAM: " + data.Encode())

	// いざfetch
	client := urlfetch.Client(c)
	resp, err := client.PostForm(posturl, data)
	if err != nil {
		// fetch失敗
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Bodyを文字列化
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()

	// Content-Typeを取得
	var ctype string
	ctypes := resp.Header["Content-Type"]
	if ctypes != nil && len(ctypes) > 0 {
		ctype = ctypes[0]
		c.Infof("ContentType: " + ctype)
	} else {
		c.Infof("ContentType: no data")
	}

	// 出力
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", ctype)
	fmt.Fprint(w, s)
}

func ProxyURL2(c appengine.Context, w http.ResponseWriter, posturl string, r *http.Request) {

	// Bodyを文字列化
	inbuf := new(bytes.Buffer)
	inbuf.ReadFrom(r.Body)
	params := inbuf.String()

	bodyType, _ := getContentType(r.Header)

	c.Infof("proxy\n url : %s\n bodytype : %s\n params : %s", posturl, bodyType, params)

	// いざfetch
	client := urlfetch.Client(c)
	resp, err := client.Post(posturl, bodyType, strings.NewReader(params))
	if err != nil {
		// fetch失敗
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Bodyを文字列化
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()

	// Content-Typeを取得
	ctype, cerr := getContentType(resp.Header)
	if cerr != nil {
		c.Infof("Responce: ContentType: no data")
	} else {
		c.Infof("Responce: ContentType: " + ctype)
	}

	// 出力
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", ctype)
	fmt.Fprint(w, s)
}

func getContentType(r http.Header) (string, error) {

	ctypes := r["Content-Type"]
	if ctypes != nil && len(ctypes) > 0 {
		return ctypes[0], nil
	} else {
		return "", errors.New("nothing")
	}
}

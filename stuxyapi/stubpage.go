package stuxy

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func getStubList(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	switch r.Method {
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Connection", "close")
		return

	case "GET":
		// datastoreから取得
		q := datastore.NewQuery("StubPage").Order("__key__")
		// Data取得
		var records []*StubPageModel
		_, err := q.GetAll(c, &records)
		if err != nil {
			http.Error(w, "could not query from datastore", http.StatusInternalServerError)
			return
		}
		// Map化
		ma := make([]map[string]interface{}, len(records))
		for i, d := range records {
			ma[i] = d.toMap()
		}
		// JSON化
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		o, _ := json.Marshal(ma)
		fmt.Fprint(w, string(o))
	}
}

func GetStubPage(c appengine.Context, path string) (*StubPageModel, error) {

	// URLからKeyを作成する
	key := datastore.NewKey(c, "StubPage", path, 0, nil)

	// datastoreから取得
	q := datastore.NewQuery("StubPage").Filter("__key__ =", key)

	var records []*StubPageModel
	keys, err := q.GetAll(c, &records)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, nil
	}

	return records[0], nil
}

//--------------------------------------------------------

type StubPageModel struct {

	// As Key too
	Path string
	// Title
	Title string `datastore:",noindex"`
	// Content-Type
	ContentType string `datastore:",noindex"`
	// Data
	Data string `datastore:",noindex"`
	// Mode
	Mode string `datastore:",noindex"`
	// Wait
	Wait int `datastore:",noindex"`
	// Group
	Group string
	//
	Lastupdate time.Time
}

func (model *StubPageModel) Key(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "StubPage", model.Path, 0, nil)
}

/** 自身の内容で保存 */
func PutToDataStore(c appengine.Context, model *StubPageModel) error {

	// Keyが設定されていなければ新しいキーを設定する
	key := model.Key(c)

	// LastUpdateを更新
	model.Lastupdate = time.Now()

	// datastoreに保存
	_, err := datastore.Put(c, key, model)
	if err != nil {
		c.Infof("Put Error:%s", err)
		return err
	}

	return nil
}

/** ShopをMapデータに変換 for JSon */
func (act *StubPageModel) toMap() map[string]interface{} {

	res := make(map[string]interface{})

	res["title"] = act.Title
	res["path"] = act.Path
	res["contenttype"] = act.ContentType
	res["data"] = act.Data
	res["mode"] = act.Mode
	res["wait"] = act.Wait
	res["group"] = act.Group
	res["lastupdate"] = FormatJST(act.Lastupdate, time.RFC3339)

	return res
}

/** JST */
var JST *time.Location = time.FixedZone("Asia/Tokyo", 9*60*60)

func FormatJST(t time.Time, f string) string {
	return t.In(JST).Format(f)
}

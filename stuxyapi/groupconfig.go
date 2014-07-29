package stuxy

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func handleConfigList(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	switch r.Method {
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Connection", "close")
		return

	case "GET":
		// datastoreから取得
		q := datastore.NewQuery("GroupConfig").Order("__key__")
		// Data取得
		var records []*GroupConfigModel
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

/** Configデータを返す */
func handleConfig(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	r.ParseForm()
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.Method {
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Connection", "close")
		return
	case "GET":
		group := r.FormValue("group")
		if g, err := getGroupConfigFromDataStore(c, group); err != nil {
			http.Error(w, "could not put to datastore", http.StatusBadRequest)
			return
		} else {
			// Map化
			ma := make([]map[string]interface{}, 1)
			ma[0] = g.toMap()
			// JSON化
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			o, _ := json.Marshal(ma)
			fmt.Fprint(w, string(o))
		}

	case "POST":
		group := r.FormValue("group")
		proxyurl := r.FormValue("proxyurl")

		model := GroupConfigModel{
			Group:    group,
			ProxyURL: proxyurl,
		}
		if err := model.PutToDataStore(c); err != nil {
			http.Error(w, "could not put to datastore", http.StatusInternalServerError)
			return
		}

		// Memcached
		if err := model.PutToMemcache(c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "DELETE":
		group := r.FormValue("group")

		config, err := getGroupConfigFromDataStore(c, group)
		if err == nil {
			if err2 := datastore.Delete(c, config.Key(c)); err2 != nil {
				c.Errorf("can't delete:%s", err2)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			c.Infof("no group " + group)
		}
	}
}

func getGroupConfigFromDataStore(c appengine.Context, group string) (*GroupConfigModel, error) {

	// URLからKeyを作成する
	key := datastore.NewKey(c, "GroupConfig", group, 0, nil)

	// datastoreから取得
	q := datastore.NewQuery("GroupConfig").Filter("__key__ =", key)

	var records []*GroupConfigModel
	keys, err := q.GetAll(c, &records)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, nil
	}

	return records[0], nil
}

func LoadGroupConfig(c appengine.Context, group string) (*GroupConfigModel, error) {

	// memcacheから取得してみる
	item, err := memcache.Get(c, "GroupConfig."+group)
	if err == nil {
		var v GroupConfigModel
		if err := json.Unmarshal(item.Value, &v); err == nil {
			return &v, nil
		}

	}

	// memcahceにないのでDBからload
	v, err := getGroupConfigFromDataStore(c, group)
	if err != nil {
		return nil, err
	}

	// memcachedにPut
	if v != nil {
		v.PutToMemcache(c)
		return v, nil
	} else {
		return nil, nil
	}
}

//--------------------------------------------------------

type GroupConfigModel struct {

	// As Key too
	Group string
	// Title
	ProxyURL string `datastore:",noindex"`
	//
	Lastupdate time.Time
}

func (model *GroupConfigModel) PutToMemcache(c appengine.Context) error {
	// GroupのJSon化
	j, err := json.Marshal(model)
	if err != nil {
		return err
	}
	c.Infof(string(j))

	// キャッシュを書き換え
	item := &memcache.Item{
		Key:        "GroupConfig." + model.Group,
		Value:      j,
		Expiration: time.Duration(5) * time.Minute,
	}
	memcache.Set(c, item)
	return nil
}

func (model *GroupConfigModel) Key(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "GroupConfig", model.Group, 0, nil)
}

/** 自身の内容で保存 */
func (model *GroupConfigModel) PutToDataStore(c appengine.Context) error {

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
func (act *GroupConfigModel) toMap() map[string]interface{} {

	res := make(map[string]interface{})

	res["group"] = act.Group
	res["proxyurl"] = act.ProxyURL
	res["lastupdate"] = FormatJST(act.Lastupdate, time.RFC3339)

	return res
}

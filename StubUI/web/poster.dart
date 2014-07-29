import 'package:polymer/polymer.dart';
import 'dart:html';
import 'dart:convert';

@CustomTag('post-form')
class Poster extends PolymerElement {
  
  @published String  hosturl = "";
  
  @observable String  errorMessage = null;
  
  @observable String  formFQDN;
  @observable String  formPath;
  @observable String  formParam;
  
  @observable String  result = null;

  Poster.created() : super.created() {
    clearForm();
  }
  
  /** フォームの初期化 */
  void clearForm() {
    
    // 入力内容を保存
    var last = window.localStorage["post-form-LastData"];
    if( last!=null ) {
      print("loaded data="+last);
      var map = JSON.decode(last);
      formFQDN = map["fqdn"];
      formPath = map["path"];
      formParam = map["param"];
      
    } else {
      
      formFQDN = "http://localhost:8081";
      formPath = "/test";
      formParam = "";
    }
    
  }
  
  /** フォームの呼び出し */
  void sendForm() {
    
    var data = new Map();
    // パラメーターをMap化
    RegExp exp = new RegExp(r"^(\w+)=(.*)$");
    formParam.split("\n").forEach( (v) {
      var m = exp.firstMatch(v);
      if( m!=null ) {
        data[m[1]] = m[2];
      }
    });
    
    var param = {
      "url" : "$formFQDN$formPath",
      "data" : encodeMap(data),
    };
    
    
    print("senddata=$data");
    // 送信
    HttpRequest.postFormData("$hosturl/env/fetch/postform", param)
    .then( (HttpRequest req) {
      errorMessage = null;
      // 受信データでメンバ書き換え
      result = req.responseText;
      // 入力内容を保存
      window.localStorage["post-form-LastData"] = JSON.encode({
        "fqdn" : formFQDN,
        "path" : formPath,
        "param" : formParam,
      });

    })
    .catchError((e){
      try {
        HttpRequest r = e.target;
        errorMessage = "Error!: readyState=${r.readyState}, status=${r.status}, response=${r.responseText}";
      } catch(e) {
        print(e);
        errorMessage = "Unknown Error!: ${e.toString()}";
      }
      result = null;
    });
  }
  
}

String encodeMap(Map data) {
  return data.keys.map((k) {
    return '${Uri.encodeComponent(k)}=${Uri.encodeComponent(data[k])}';
  }).join('&');
}

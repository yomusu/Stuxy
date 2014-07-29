import 'package:polymer/polymer.dart';
import 'dart:html';
import 'dart:convert';

@CustomTag('stub-list')
class StubList extends PolymerElement {
  
  /** Waitの秒数Selectの選択肢 */
  @published List<Map> selectWaitData = [
     {"title":"なし", "value": 0 },
     {"title":"5秒", "value": 5 },
     {"title":"15秒", "value": 15 },
     {"title":"30秒", "value": 30 },
  ];
  
  /** ReturnTypeのSelectの選択肢 */
  @published List<Map> selectModeData = [
     {"title":"Stub", "value": "MDSTUB" },
     {"title":"Proxy", "value": "MDPROXY" },
     {"title":"404", "value": "MD404" },
  ];
  
  
  @published  String  hosturl = "";
  
  @observable String  errorMessage = null;
  
  @observable List  datas = null;

  StubList.created() : super.created() {
    updateData();
    clearForm();
  }
  
  /** スタブデータ一覧の読み込み */
  void updateData() {
    
    HttpRequest.getString("${hosturl}/env/list")
    .then( (str) {
      // JSON文字列をMap化
      var all = JSON.decode(str) as List;
      datas = all;
      // Waitをindex化
      var valList = new List.from(selectWaitData.map( (e)=>e['value'] ),growable: false);
      datas.forEach( (v){
        var i = valList.indexOf(v['wait']);
        v['waitIndex'] = (i>=0) ? i : 0;
      });
      // RetTypeをindex化
      var retlist = new List.from(selectModeData.map( (e)=>e['value'] ),growable: false);
      datas.forEach( (v){
        var i = retlist.indexOf(v['mode']);
        v['modeIndex'] = (i>=0) ? i : 0;
      });
    })
    .catchError((e){ errorMessage = "StubListの読み込みに失敗しました。e=${e}"; });
  }
  
  /** スタブひとつを削除 */
  void delStub(Event e, var detail, Element target ) {
    var path = target.attributes['stub-path'];
    var stub = _searchStub(path);
    
    // 送信
    HttpRequest.request("${hosturl}/env/post?path=${stub['path']}",method:"DELETE")
    .then( (HttpRequest req) { updateData(); })
    .catchError((e){ errorMessage = "${stub['path']}の削除に失敗しました。e=${e}"; });
  }
  
  /** スタブひとつを編集開始 */
  void editStub(Event e, var detail, Element target ) {
    var path = target.attributes['stub-path'];
    var stub = _searchStub(path);
    
    formTitle = stub['title'];
    formPath = stub['path'];
    formContentType = stub['contenttype'];
    formData = stub['data'];
    formGroup = stub['group'];
  }
  
  /** スタブを検索 */
  Map _searchStub( String path ) => datas.firstWhere( (v) => v['path']==path );
  
  @observable String  formTitle;
  @observable String  formPath;
  @observable String  formContentType;
  @observable String  formData;
  @observable String  formGroup;
  
  /** フォームの初期化 */
  void clearForm() {
    formTitle = "";
    formPath = "/";
    formContentType = "text/html; charset=UTF-8";
    formData = "";
    formGroup = "";
  }
  
  void sendForm() {
    var data = {
                "title" : formTitle,
                "path"  : formPath,
                "contenttype": formContentType,
                "data"  : formData,
                "group" : formGroup,
    };
    HttpRequest.postFormData("${hosturl}/env/post", data)
    .then( (HttpRequest req) { updateData(); })
    .catchError((e){ errorMessage = "Stubの送信に失敗しました。e=${e}"; });
  }
  
  /** Waitの値が変更されたのでサーバーに通知 */
  void changeWaitSelection(Event e, var detail, Element target ) {
    var path = target.attributes['stub-path'];
    var stub = _searchStub(path);
    
    var i = stub['waitIndex'];
    
    print("wait is updated to ${selectWaitData[i]}");
    HttpRequest.postFormData("${hosturl}/env/mode", {
      "path": path,
      "wait" : selectWaitData[i]['value'].toString(),
    })
    .then( (HttpRequest req) {})
    .catchError((e){ errorMessage = "Waitの送信に失敗しました。e=${e}"; });
  }
  
  /** Modeの値が変更されたのでサーバーに通知 */
  void changeModeSelection(Event e, var detail, Element target ) {
    var path = target.attributes['stub-path'];
    var stub = _searchStub(path);
    var i = stub['modeIndex'];
    
    HttpRequest.postFormData("${hosturl}/env/mode", {
      "path": path,
      "mode" : selectModeData[i]['value'].toString(),
    })
    .then( (HttpRequest req) { })
    .catchError((e){ errorMessage = "Waitの送信に失敗しました。e=${e}"; });
  }
}


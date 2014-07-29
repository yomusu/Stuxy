import 'package:polymer/polymer.dart';
import 'dart:html';
import 'dart:convert';

@CustomTag('group-list')
class GroupList extends PolymerElement {
  
  @published  String  hosturl = "";
  
  @observable String  errorMessage = null;
  
  @observable List  datas = null;

  GroupList.created() : super.created() {
    updateData();
    clearForm();
  }
  
  /** Groupデータ一覧の読み込み */
  void updateData() {
    
    HttpRequest.getString("${hosturl}/env/config/list")
    .then( (str) {
      // JSON文字列をMap化
      var all = JSON.decode(str) as List;
      datas = all;
    })
    .catchError((e){ errorMessage = "GroupListの読み込みに失敗しました。e=${e}"; });
  }
  
  /** Groupひとつを削除 */
  void delGroup(Event e, var detail, Element target ) {
    var path = target.attributes['group-name'];
    var group = _searchGroup(path);
    
    // 送信
    HttpRequest.request("${hosturl}/env/config?group=${group['group']}",method:"DELETE")
    .then( (HttpRequest req) { updateData(); })
    .catchError((e){ errorMessage = "${group['group']}の削除に失敗しました。e=${e}"; });
  }
  
  /** スタブひとつを編集開始 */
  void editGroup(Event e, var detail, Element target ) {
    var path = target.attributes['group-name'];
    var group = _searchGroup(path);
    
    formProxyURL = group['proxyurl'];
    formGroup = group['group'];
  }
  
  /** スタブを検索 */
  Map _searchGroup( String group ) => datas.firstWhere( (v) => v['group']==group );
  
  @observable String  formProxyURL;
  @observable String  formGroup;
  
  /** フォームの初期化 */
  void clearForm() {
    formProxyURL= "";
    formGroup = "";
  }
  
  /** フォームの送信 */
  void sendForm() {
    var data = {
                "proxyurl" : formProxyURL,
                "group" : formGroup,
    };
    HttpRequest.postFormData("${hosturl}/env/config", data)
    .then( (HttpRequest req) { updateData(); })
    .catchError((e){ errorMessage = "Groupの送信に失敗しました。e=${e}"; });
  }
}


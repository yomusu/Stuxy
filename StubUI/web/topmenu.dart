import 'package:polymer/polymer.dart';
import 'dart:html';

@CustomTag('top-menu')
class TopMenu extends PolymerElement {
  
  @observable String  mode = "stub";

  TopMenu.created() : super.created() {
  }
  
  void changeMode(Event e, var detail, Element target ) {
    mode = target.attributes['mode-code'];
  }
}

import 'package:flutter/foundation.dart' show ChangeNotifier;

class LoadingViewModel with ChangeNotifier {
  int _loading = 0;

  bool get isLoading => _loading > 0;

  set isLoading(bool isLoading) {
    if (isLoading) {
      _loading++;
    } else {
      _loading--;
    }
    if (_loading == 0) {
      notifyListeners();
    }
  }
}

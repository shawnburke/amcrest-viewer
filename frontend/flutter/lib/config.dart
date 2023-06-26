import 'package:flutter/foundation.dart';

class Config {
  static String? _baseURL;
  static String get baseURL {
    _baseURL ??= const String.fromEnvironment('BASE_URL', defaultValue: '');

    if (_baseURL!.endsWith('/')) {
      _baseURL = _baseURL!.substring(0, _baseURL!.length - 1);
    }
    return _baseURL!;
  }
}

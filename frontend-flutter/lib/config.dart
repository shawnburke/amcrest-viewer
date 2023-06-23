// const String? _urlOverride = "http://0.0.0.0:3001";
//const String? _urlOverride = "http://192.168.86.10:9000";

class Config {
  static String? _baseURL;
  static String get baseURL {
    // todo load assets
    // if (_urlOverride != null) {
    //   return _urlOverride!;
    // }
    _baseURL ??= const String.fromEnvironment('BASE_URL', defaultValue: '');

    if (_baseURL!.endsWith('/')) {
      _baseURL = _baseURL!.substring(0, _baseURL!.length - 1);
    }
    return _baseURL!;
  }
}
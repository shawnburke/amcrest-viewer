import 'package:dio/dio.dart';
import 'package:logging/logging.dart';

abstract class RequestCache {
  bool get enabled => true;
  Future<Response?> fetch(RequestOptions req);
  Future<void> save(Response res);
}

class MemoryCache implements RequestCache {
  final log = Logger('MemoryCache');

  String getRequestKey(RequestOptions options) {
    var key = '${options.method}+${options.baseUrl}${options.path}';
    var query = "";
    for (final e in options.queryParameters.entries) {
      query += "${e.key}=${e.value}&";
    }
    if (query.isNotEmpty) {
      key = "$key?$query";
    }
    if (options.data != null) {
      key = "$key|data=${options.data.hashCode}";
    }
    return key;
  }

  final cache = <String, Response>{};

  @override
  Future<Response?> fetch(RequestOptions options) async {
    final key = getRequestKey(options);
    if (cache.containsKey(key)) {
      log.fine('Cache hit: $key');
      final cached = cache[key]!;
      final header = cached.headers.value('x-cache-replay');
      if (header == null) {
        cached.headers.add('x-cache-replay', 'true');
      }
      return cached;
    }
    return null;
  }

  @override
  Future<void> save(Response response) async {
    final info = <String, dynamic>{
      'statusCode': response.statusCode,
      'statusMessage': response.statusMessage,
      'headers': response.headers,
      'data': response.data,
    };
    log.info('Response: $info');
    final key = getRequestKey(response.requestOptions);
    cache[key] = response;
  }

  @override
  bool get enabled => true;
}

import 'package:dio/dio.dart';

class TrafficInterceptor extends Interceptor {
  final cache = <String, Response>{};

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

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final info = <String, dynamic>{
      'uri':
          '${options.method} ${options.baseUrl}${options.path}?${options.queryParameters}',
      'method': options.method,
      'path': options.path,
      'query': options.queryParameters,
      'headers': options.headers,
      'data': options.data,
    };
    final key = getRequestKey(options);
    if (cache.containsKey(key)) {
      print('Cache hit: $key');
      final cached = cache[key]!;
      final header = cached.headers.value('x-cache-replay');
      if (header == null) {
        cached.headers.add('x-cache-replay', 'true');
      }
      handler.resolve(cached);
      return;
    }
    print('Request: $info');
    super.onRequest(options, handler);
  }

  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    final info = <String, dynamic>{
      'statusCode': response.statusCode,
      'statusMessage': response.statusMessage,
      'headers': response.headers,
      'data': response.data,
    };
    print('Response: $info');
    final key = getRequestKey(response.requestOptions);
    cache[key] = response;
    super.onResponse(response, handler);
  }

  @override
  void onError(DioError err, ErrorInterceptorHandler handler) {
    super.onError(err, handler);
  }
}

import 'dart:indexed_db';

import 'package:dio/dio.dart';

import 'cache.dart';
import 'mockoon_cache.dart';

class TrafficInterceptor extends Interceptor {
  late final RequestCache cache;

  TrafficInterceptor({RequestCache? cache, String? baseUrl}) {
    this.cache = cache ?? MockoonCache(baseUrl ?? 'http://localhost:8080');
  }

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    cache.fetch(options).then((response) {
      if (response != null) {
        handler.resolve(response);
        return;
      }
      super.onRequest(options, handler);
    }).onError((error, stackTrace) {
      print('Error: $error');
      super.onRequest(options, handler);
    });
  }

  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    if (response.headers.value('x-cache-replay') == null) {
      cache.save(response);
    }
    super.onResponse(response, handler);
  }

  @override
  void onError(DioError err, ErrorInterceptorHandler handler) {
    print('Error: $err');
    super.onError(err, handler);
  }
}

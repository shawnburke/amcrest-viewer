import 'package:dio/dio.dart';

import 'cache.dart';

class TrafficInterceptor extends Interceptor {
  late final RequestCache cache;

  TrafficInterceptor(this.cache);

  bool get enabled => cache.enabled;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    if (!enabled) {
      super.onRequest(options, handler);
      return;
    }

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

    if (!enabled) {
      super.onResponse(response, handler);
      return;
    }

    if (response.headers.value('x-cache-replay') == null) {
      cache.save(response);
    }
    super.onResponse(response, handler);
  }

  @override
  void onError(DioError err, ErrorInterceptorHandler handler) {

    if (!enabled) {
      super.onError(err, handler);
      return;
    }

    print('Error: $err');
    super.onError(err, handler);
  }
}

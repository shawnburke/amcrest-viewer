import 'package:dio/dio.dart';
import 'package:logging/logging.dart';

import 'cache.dart';

class TrafficInterceptor extends Interceptor {
  final log = Logger('TrafficInterceptor');

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
      log.severe('$error');
      super.onRequest(options, handler);
    });
  }

  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    if (!enabled) {
      super.onResponse(response, handler);
      return;
    }

    // never save cache replay responses
    if (response.headers.value('x-cache-replay') != null) {
      super.onResponse(response, handler);
      return;
    }

    cache.save(response).then((r) {
      super.onResponse(response, handler);
    }).onError((error, stackTrace) {
      log.severe('$error');
      super.onResponse(response, handler);
    });
  }

  @override
  void onError(DioError err, ErrorInterceptorHandler handler) {
    if (!enabled) {
      super.onError(err, handler);
      return;
    }

    ;
    log.severe('$err');
    super.onError(err, handler);
  }
}

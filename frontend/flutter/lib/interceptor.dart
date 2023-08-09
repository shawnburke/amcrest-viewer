import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:mockoon_proxy/request_info.dart';
import 'package:mockoon_proxy/response_info.dart';

class TrafficInterceptor extends Interceptor {
  late final RequestCache cache;

  TrafficInterceptor() {
    cache = MockoonCache('http://localhost:8080');
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

  // @override
  // void onError(DioError err, ErrorInterceptorHandler handler) {
  //   super.onError(err, handler);
  // }
}

abstract class RequestCache {
  Future<Response?> fetch(RequestOptions req);
  Future<void> save(Response res);
}

class MemoryCache implements RequestCache {
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
    // final info = <String, dynamic>{
    //   'uri':
    //       '${options.method} ${options.baseUrl}${options.path}?${options.queryParameters}',
    //   'method': options.method,
    //   'path': options.path,
    //   'query': options.queryParameters,
    //   'headers': options.headers,
    //   'data': options.data,
    // };
    final key = getRequestKey(options);
    if (cache.containsKey(key)) {
      print('Cache hit: $key');
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
    print('Response: $info');
    final key = getRequestKey(response.requestOptions);
    cache[key] = response;
  }
}

class MockoonCache implements RequestCache {
  late final Dio dio;

  MockoonCache(String baseUrl) {
    if (!baseUrl.endsWith('/')) {
      baseUrl += '/';
    }
    dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      headers: {
        'Content-Type': 'application/json',
      },
    ));
  }

  @override
  Future<Response?> fetch(RequestOptions req) async {
    final start = DateTime.now();
    final fetchResponse = await dio.post(
      'scenarios/test_scenario/fetch',
      data: RequestInfo.fromDio(req).toJson(),
      options: Options(
        validateStatus: (status) => true,
        receiveDataWhenStatusError: true,
      ),
    );

    if (fetchResponse.statusCode == 200) {
      print(
          'Cache hit: ${req.uri}, took ${DateTime.now().difference(start).inMilliseconds}ms');
      final res = Response(
        requestOptions: req,
        statusCode: fetchResponse.statusCode,
        statusMessage: '',
        headers: fetchResponse.headers,
        data: fetchResponse.data,
      );
      return res;
    }
    print(
        'Cache miss: ${req.uri}, took ${DateTime.now().difference(start).inMilliseconds}ms');
    return null;
  }

  @override
  Future<void> save(Response res) async {
    final resInfo = ResponseInfo.fromDio(res);
    final serverResponse =
        await dio.post('scenarios/test_scenario/save', data: resInfo.toJson());
    print('Server response: $serverResponse');
  }
}

import 'dart:io';

import 'package:dio/dio.dart';
import 'package:logging/logging.dart';
import 'package:mockoon_proxy/models/request_info.dart';
import 'package:mockoon_proxy/models/response_info.dart';

import 'cache.dart';
import 'interceptor.dart';

class MockoonProxyClient implements RequestCache {
  final log = Logger('MockDataServerClient');
  static const String serverEnvVar = 'MOCKOON_PROXY_HOSTPORT';

  // if the proxy isn't running, we don't want to timeout
  // on every request so if we get a socket timeout, we ignore and try again
  //
  DateTime? disabledUntil;
  Duration disabledTimeout;
  Duration connectTimeout;

  late final Dio _dio;

  @override
  bool get enabled {
    if (_dio.options.baseUrl.isEmpty) {
      return false;
    }

    if (_isDisabled) {
      return false;
    }

    return true;
  }

  MockoonProxyClient({
    String? mockoonProxyHostPort,
    this.disabledTimeout = const Duration(seconds: 30),
    this.connectTimeout = const Duration(milliseconds: 100),
  }) {
    mockoonProxyHostPort ??=
        const String.fromEnvironment(serverEnvVar, defaultValue: '');

    // if we have one, we set it up and enable the cache.
    var baseUrl = '';
    if (mockoonProxyHostPort.isNotEmpty) {
      baseUrl = "http://$mockoonProxyHostPort/";
    }

    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      headers: {
        'Content-Type': 'application/json',
      },
      connectTimeout: connectTimeout,
      validateStatus: (status) => true,
      receiveDataWhenStatusError: true,
    ));
  }

  @override
  Future<Response?> fetch(RequestOptions req) async {
    if (!enabled) {
      return null;
    }

    try {
      final start = DateTime.now();
      final fetchResponse = await _dio.post(
        'scenarios/current/fetch',
        data: RequestInfo.fromDio(req).toJson(),
      );

      switch (fetchResponse.statusCode) {
        case 200:
          log.fine(
              'Cache hit: ${req.uri}, took ${DateTime.now().difference(start).inMilliseconds}ms');
          final res = Response(
            requestOptions: req,
            statusCode: fetchResponse.statusCode,
            statusMessage: '',
            headers: fetchResponse.headers,
            data: fetchResponse.data,
          );
          return res;
        case 418:
          // server is disabled or wants us to fall through
          return null;

        case 404:
        case 501:
          // server is enabled but has no response
          // so client should fail
          log.warning(
              "no mockoon response available for ${req.method} ${req.uri}, it may have not been recorded?");
          return fetchResponse;
      }
      log.warning(
          'Cache miss: ${req.uri}, took ${DateTime.now().difference(start).inMilliseconds}ms');
    } on DioException catch (e) {
      final se = e.error as SocketException?;
      if (se == null) {
        log.shout('Unexpected DIO error: ${e.error}');
        rethrow;
      }
      _disable(se);
    }
    return null;
  }

  @override
  Future<void> save(Response res) async {
    if (!enabled) {
      return;
    }
    final resInfo = ResponseInfo.fromDio(res);
    await _dio.post('scenarios/current/save', data: resInfo.toJson());
  }

  void _disable(SocketException e) {
    if (!enabled) {
      return;
    }
    log.severe(
        'Cannot contact mockoon-proxy (${e.message}): disabling interceptor for $disabledTimeout');
    disabledUntil = DateTime.now().add(disabledTimeout);
  }

  bool get _isDisabled {
    if (disabledUntil?.isAfter(DateTime.now()) ?? false) {
      return true;
    }
    disabledUntil = null;
    return false;
  }
}

class MockoonProxyInterceptor extends TrafficInterceptor {
  MockoonProxyInterceptor({int? port})
      : super(MockoonProxyClient(
            mockoonProxyHostPort: port == null ? null : 'localhost:$port'));
}

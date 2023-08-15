import 'package:dio/dio.dart';
import 'package:mockoon_proxy/models/request_info.dart';
import 'package:mockoon_proxy/models/response_info.dart';

import 'cache.dart';

class MockoonProxyClient implements RequestCache {

  static const String serverEnvVar = 'MOCKOON_PROXY_HOSTPORT';

  late final Dio _dio;

  @override
  late final bool enabled;
  
  MockoonProxyClient({String? mockoonProxyHostPort}) {

    // if no hostport is passed in we look for an environment variable
    if (mockoonProxyHostPort == null) {
        final envVar = String.fromEnvironment(serverEnvVar);
        if (envVar.isNotEmpty) {
          mockoonProxyHostPort = envVar;
        }
    }

    // if we have one, we set it up and enable the cache.
    if (mockoonProxyHostPort != null) {
      _dio = Dio(BaseOptions(
        baseUrl: "http://$mockoonProxyHostPort",
        headers: {
          'Content-Type': 'application/json',
          
        },
        validateStatus: (status) => true,
        receiveDataWhenStatusError: true,
      ));
      enabled = true;
    } else {
      _dio = Dio();
    }
  }

  @override
  Future<Response?> fetch(RequestOptions req) async {
    final start = DateTime.now();
    final fetchResponse = await _dio.post(
      'scenarios/current/fetch',
      data: RequestInfo.fromDio(req).toJson(),
    );

    switch (fetchResponse.statusCode) {
      case 200:
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
      case 418:
        // server is disabled or wants us to fall through
        return null;
      case 501:
        // server is enabled but has no response
        // so client should fail
        print("Error: no response available for ${req.method} ${req.uri}");
        return fetchResponse;
    }
    print(
        'Cache miss: ${req.uri}, took ${DateTime.now().difference(start).inMilliseconds}ms');
    return null;
  }

  @override
  Future<void> save(Response res) async {
    final resInfo = ResponseInfo.fromDio(res);
    final serverResponse =
        await _dio.post('scenarios/current/save', data: resInfo.toJson());

    if (serverResponse.statusCode == 418) {
      // server disabled
      return null;
    }
    print('Server response: $serverResponse');
  }
}

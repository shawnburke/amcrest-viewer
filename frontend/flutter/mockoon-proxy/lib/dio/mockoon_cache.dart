import 'package:dio/dio.dart';
import 'package:mockoon_proxy/request_info.dart';
import 'package:mockoon_proxy/response_info.dart';

import 'cache.dart';

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

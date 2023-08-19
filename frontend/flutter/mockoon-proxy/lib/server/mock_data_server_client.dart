import 'package:dio/dio.dart' as dio;
import 'package:mockoon_proxy/server/mock_data_server.dart';

class MockoonServerClient {
  final int port;
  final String host;
  late final dio.Dio _dio;

  MockoonServerClient(this.port, {this.host = 'localhost'}) {
    _dio = dio.Dio(
      dio.BaseOptions(
        baseUrl: 'http://$host:$port',
        receiveDataWhenStatusError: true,
        validateStatus: (status) => true,
      ),
    );
  }

  Map<String, dynamic> _toDictionary(dio.Response response) {
    final Map<String, dynamic> res = {};
    dynamic data = response.data;
    if (data is Map<String, dynamic>) {
      res.addAll(data);
    } else if (data is String) {
      res["body"] = data;
      print('NOT JSON -- body: $data');
    }
    res["headers"] = response.headers;
    return res;
  }

  Future<Map<String, dynamic>> setMode(
      {Mode mode = Mode.disabled, String? scenario}) async {
    var path = '/mode/${mode.name}';
    if (scenario != null) {
      path += '?scenario=$scenario';
    }
    final response = await _dio.post(path);
    return _toDictionary(response);
  }

  Future<List<String>> getScenarios() async {
    final response = await _dio.get('/scenarios');
    final List<dynamic> l = response.data['scenarios'];

    return l.cast<String>();
  }

  Future<Map<String, dynamic>> getMode() async {
    final response = await _dio.get('/mode');
    return _toDictionary(response);
  }

  Future<Map<String, dynamic>> startScenario(
      {String scenario = 'current'}) async {
    final response = await _dio.post('/scenarios/$scenario/start');
    return _toDictionary(response);
  }

  Future<Map<String, dynamic>> closeScenario(
      {String scenario = 'current'}) async {
    final response = await _dio.post('/scenarios/$scenario/close');
    return _toDictionary(response);
  }
}

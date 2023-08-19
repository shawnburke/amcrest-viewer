import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart' as dio;

import 'package:mockoon_proxy/dio/interceptor.dart';
import 'package:mockoon_proxy/dio/mockoon_proxy_client.dart';
import 'package:mockoon_proxy/models/response_info.dart';
import 'package:mockoon_proxy/server/mock_data_server.dart';
import 'package:mockoon_proxy/server/mock_data_server_client.dart';
import 'package:shelf/shelf.dart';
import 'package:shelf_router/shelf_router.dart';
import 'package:shelf/shelf_io.dart' as io;

import 'package:test/test.dart';

void main() {
  test('e2e', () async {
    final server = _TestServer();

    final serverPort = await server.run();

    // final tmpDir = Directory.systemTemp.path +
    //     '/mockoon_proxy_test/${DateTime.now().millisecondsSinceEpoch}';

    // for debugging
    final tmpDir = 'test/tmp';

    final mockoonServer = MockDataServer(0, dir: tmpDir, useTempDir: false);

    final mockoonPort = await mockoonServer.start();

    mockoonServer.dispose();

    try {
      final ms = MockoonServerClient(mockoonPort);

      final interceptor = MockoonProxyInterceptor(port: mockoonPort);

      expect(interceptor.enabled, true);

      final client = _TestClient(serverPort, interceptor);

      final getResponse = await client.doGet();

      expect(getResponse['val'], 123);

      final postResp = await client.doPost('dog', 'baz', 456);

      expect(postResp['category'], 'dog');

      final startMode = await ms.getMode();

      expect(startMode['mode'], 'disabled');

      // start a scenario
      final scenario = 'scenario_${DateTime.now().millisecondsSinceEpoch}';
      final startScenario =
          await ms.setMode(mode: Mode.record, scenario: scenario);
      expect(startScenario['mode'], 'record');

      const extraVal = 'ex1';
      final gr = await client.doGet();
      expect(gr.length > 0, true);
      final pr = await client.doPost('bird', 'test', 111, extra: extraVal);
      expect(pr.length > 0, true);

      final encodedVal = '[12345, 67890]';
      final pr2 = await client.doPost('fancy', 'stuff', 777, extra: encodedVal);
      expect(pr2.length > 0, true);

      // save it
      await ms.closeScenario();

      // shut down backing server
      await server.dispose();

      // now replay
      final replayScenario =
          await ms.setMode(mode: Mode.replay, scenario: scenario);
      expect(replayScenario['mode'], 'replay');

      final getResponse2 = await client.doGet();

      expect(getResponse2['val'], 123);

      final postResp2 =
          await client.doPost('bird', 'test', 111, extra: extraVal);

      expect(postResp2.containsKey('category'), true,
          reason: 'postResp2: $postResp2');
      expect(postResp2['category'], 'bird');
      expect(postResp2['extra'], extraVal);

      final postResp3 =
          await client.doPost('fancy', 'stuff', 777, extra: encodedVal);

      expect(postResp3.containsKey('category'), true,
          reason: 'postResp3: $postResp2');
      expect(postResp3['category'], 'fancy');
      expect(postResp3['extra'], encodedVal);

      // mustate the body to ensure it does not replay
      final postBadBody =
          await client.doPost('bird', 'baz', 456, extra: extraVal);

      expect(postBadBody["statusCode"], 501);

      final postBadParam =
          await client.doPost('bird', 'baz', 111, extra: 'xyz');
      expect(postBadParam["statusCode"], 501);

      // restart test server to test fallthru
      await server.run();

      final fallthruScenario =
          await ms.setMode(mode: Mode.replay_fallthrough, scenario: scenario);
      expect(fallthruScenario['mode'], Mode.replay_fallthrough.name);

      final postFallThru =
          await client.doPost('cat', 'baz', 999, extra: extraVal);

      expect(postFallThru["input_val"], 999);
    } finally {
      await server.dispose();
      await mockoonServer.dispose();

      final tmp = Directory(tmpDir);
      if (tmp.path.contains('mockoon_proxy_test')) {
        await tmp.delete(recursive: true);
      }
    }
  });

  test('Client auto disable', () async {
    final interceptor = MockoonProxyInterceptor(port: 12345);

    expect(interceptor.enabled, true);
    final result = await interceptor.cache.fetch(dio.RequestOptions());

    expect(result, isNull);
    expect(interceptor.enabled, false);
  });

  test('Clean data', () {
    final data = "{\"foo\": \"\$445,671.59\"}";

    final clean = ResponseInfo.cleanData(data);

    expect(clean, "{\"foo\": \"\$123.45\"}");
  });
}

class _TestClient {
  late final dio.Dio _dio;

  _TestClient(
    int port,
    TrafficInterceptor interceptor,
  ) {
    _dio = dio.Dio(
      dio.BaseOptions(
        baseUrl: 'http://localhost:$port',
        receiveDataWhenStatusError: true,
        validateStatus: (status) => true,
      ),
    );
    _dio.interceptors.add(interceptor);
  }

  Future<Map<String, dynamic>> doGet() async {
    final response = await _dio.get('/get');
    final result = response.data as Map<String, dynamic>;
    result["headers"] = response.headers;
    return result;
  }

  Future<Map<String, dynamic>> doPost(String cat, String foo, int val,
      {String? extra}) async {
    final body = jsonEncode({
      'foo': foo,
      'val': val,
    });

    var query = '';

    if (extra != null) {
      query = '?extra=${Uri.encodeQueryComponent(extra)}';
    }
    final Map<String, dynamic> res = {};
    try {
      final response = await _dio.post('/post/$cat$query',
          options: dio.Options(headers: {'content-type': 'application/json'}),
          data: body);

      dynamic data = response.data;
      if (data is Map<String, dynamic>) {
        res.addAll(data);
      } else if (data is String) {
        res["body"] = data;
      }
      res['statusCode'] = response.statusCode;
      res["headers"] = response.headers;
    } catch (e) {
      print('ERROR: $e');
      res['statusCode'] = 0;
      res['body'] = e.toString();
    }
    return res;
  }
}

class _TestServer {
  HttpServer? server;
  int? port;
  _TestServer();

  Future<void> dispose() async {
    if (server != null) {
      print('TestServer: shutting down');
      await server?.close();
      server = null;
    }
  }

  Future<int> run() async {
    final router = Router()
      ..get('/get', _getHandler)
      ..post('/post/<category>', _postHandler);

    var handler = Pipeline().addMiddleware(logRequests()).addHandler(router);

    // Create a server and bind it to a specific address and port.
    server = await io.serve(handler, 'localhost', port ?? 0);
    print('TestServer: running on localhost:${server!.port}');
    port = server!.port;
    return port!;
  }

  Future<Response> _getHandler(Request req) async {
    final json = JsonEncoder.withIndent(" ")
        .convert({"foo": "bar", "val": 123, "type": "get"});
    return Response.ok(json, headers: {
      'content-type': 'application/json',
      'x-type': 'get',
    });
  }

  Future<Response> _postHandler(Request req, String category) async {
    final extraVal = req.url.queryParameters['extra'];
    final body = await req.readAsString();
    final input = jsonDecode(body);
    final json = jsonEncode({
      "input_foo": input["foo"],
      "input_val": input["val"],
      "type": "get",
      "category": category,
      "extra": extraVal,
    });
    return Response.ok(json, headers: {
      'content-type': 'application/json',
      'x-type': 'get',
    });
  }
}

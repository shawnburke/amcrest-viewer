import 'dart:convert';
import 'dart:io';

import 'package:logging/logging.dart';
import 'package:mockoon_proxy/models/request_info.dart';
import 'package:mockoon_proxy/models/response_info.dart';
import 'package:mockoon_proxy/server/scenario_manager.dart';
import 'package:shelf/shelf.dart';
import 'package:shelf_router/shelf_router.dart';
import 'package:shelf/shelf_io.dart' as io;
import 'package:shelf_cors_headers/shelf_cors_headers.dart';

const serverModeHeader = 'x-mockoon-proxy-mode';

class MockDataServer {
  final log = Logger('MockDataServer');

  HttpServer? _server;
  String? currentScenario;

  final String host;
  final int port;
  final String dir;
  final bool useTempDir;
  MockDataServer(this.port,
      {this.host = '0.0.0.0', this.dir = "scenarios", this.useTempDir = true});

  Future<void> dispose() async {
    if (_server != null) {
      await _server?.close(force: true);
      _server = null;
    }

    if (manager != null) {
      manager?.dispose();
      manager = null;
    }
  }

  Future<int> start() async {
    if (_server != null) {
      return _server!.port;
    }

    final router = Router()
      ..get('/scenarios', _listHandler)
      ..get('/mode', _getModeHandler)
      ..post('/mode/<value>', _setModeHandler)
      ..post('/scenarios/<scenario>/start', _startHandler)
      ..post('/scenarios/<scenario>/save', _saveHandler)
      ..post('/scenarios/<scenario>/close', _closeHandler)
      ..post('/scenarios/<scenario>/fetch', _fetchHandler);

    // Create a pipeline that handles requests.
    var handler = Pipeline()
        .addMiddleware(logRequests(logger: (msg, isError) {
          // cut off the timestamp
          // 2023-08-17T14:38:13.775592  0:00:00.010014
          msg = msg.substring(42).trim();
          if (isError) {
            log.severe(msg);
            return;
          }
          log.fine(msg);
        }))
        .addMiddleware(corsHeaders())
        .addHandler(router);

    // Create a server and bind it to a specific address and port.
    _server = await io.serve(handler, host, port);

    log.info('Serving at http://${_server!.address.host}:${_server!.port}');
    return _server!.port;
  }

  ScenarioManager? manager;

  Mode serverMode = Mode.disabled;

  ScenarioManager getManager() {
    if (manager == null) {
      manager = ScenarioManager(dir, useTempDir: useTempDir);
    }
    return manager!;
  }

  Future<void> setScenarioPort(String scenario, int port) async {
    await getManager().setScenarioPort(scenario, port);
  }

// starts a new scenario by creating the directory
  Future<Response> _listHandler(Request request) async {
    final scenarios = await getManager().listScenarios();
    return Response.ok(
        jsonEncode({
          'scenarios': scenarios,
        }),
        headers: {
          'content-type': 'application/json',
        });
  }

  Future<Response> _getModeHandler(Request request) async {
    return Response.ok(
        jsonEncode({
          'message': 'current mode',
          'mode': serverMode.name,
          'current': currentScenario,
        }),
        headers: {
          'content-type': 'application/json',
        });
  }

  Future<Response> _setModeHandler(Request request, String value) async {
    final String? current = request.url.queryParameters['scenario'];
    if (current != null) {
      currentScenario = current;
    }

    Mode? newMode;
    for (final v in Mode.values) {
      if (v.name == value) {
        newMode = v;
        break;
      }
    }

    if (newMode == null) {
      return Response(400,
          body: jsonEncode({
            'message': 'invalid mode',
            'mode': value,
          }));
    }

    serverMode = newMode;

    return _getModeHandler(request);
  }

  String? checkCurrent(String scenario) {
    if (scenario == 'current') {
      if (currentScenario == null) {
        return null;
      }
      return currentScenario!;
    }

    return scenario;
  }

  Response noCurrentSet = Response(400,
      body: jsonEncode({
        'message': 'no current scenario set',
      }));

// starts a new scenario by creating the directory
  Future<Response> _startHandler(Request request, String scenario) async {
    final s = checkCurrent(scenario);
    if (s == null) {
      return noCurrentSet;
    }
    bool clear = request.url.queryParameters['clear'] == 'true' ? true : false;
    bool reset = await getManager().startScenario(s, clear: clear);
    return Response.ok(jsonEncode({
      'message': 'created: reset=$reset',
      'mode': serverMode.name,
      'scenario': s,
      'reset': reset,
    }));
  }

// starts a new scenario by procesing all of the files into a
// mockoon file
  Future<Response> _closeHandler(Request request, String scenario) async {
    final s = checkCurrent(scenario);
    if (s == null) {
      return noCurrentSet;
    }
    final path = await getManager().closeScenario(s);
    serverMode = Mode.disabled;
    return Response.ok(
        jsonEncode({
          'message': 'closed',
          'mode': serverMode.name,
          'scenario': s,
          'path': path,
        }),
        headers: {
          'content-type': 'application/json',
        });
  }

// saves a file into the directory
  Future<Response> _saveHandler(Request request, String scenario) async {
    switch (serverMode) {
      case Mode.disabled:
      case Mode.replay:
      case Mode.replay_fallthrough:
        // server is disabled, so client should
        // call as normal through to the real server.
        return Response(418,
            body: jsonEncode({
              'message': 'server mode is disabled',
              'mode': serverMode.name,
            }));

      default:
        break;
    }

    final s = checkCurrent(scenario);
    if (s == null) {
      return noCurrentSet;
    }
    final body = await request.readAsString();

    final json = jsonDecode(body);

    final response = ResponseInfo.fromJson(json);
    await getManager().saveResponse(s, response);
    return Response.ok(jsonEncode({
      'message': 'saved',
      'request': response.request.path,
      'scenario': s,
    }));
  }

  Future<Response> _fetchHandler(Request request, String scenario) async {
    final body = await request.readAsString();

    final json = jsonDecode(body);

    final ri = RequestInfo.fromJson(json);

    switch (serverMode) {
      case Mode.disabled:
      case Mode.record:
        log.fine(
            'Ignoring fetch request ${ri.method} ${ri.uri} in mode $serverMode');
        // server is disabled, so client should
        // call as normal through to the real server.
        return Response(418,
            body: jsonEncode({
              'message': 'server mode is disabled',
              'mode': serverMode.name,
            }));

      default:
        break;
    }

    final s = checkCurrent(scenario);
    if (s == null) {
      return noCurrentSet;
    }

    final res = await getManager().getResponse(s, ri);
    if (res == null || res.statusCode == 501) {
      if (serverMode == Mode.replay_fallthrough) {
        final info = _buildRequestInfo(ri);
        log.warning('The request ${ri.method} ${ri.uri} was not found in the '
            'scenario $s, falling through to the real server. You may need to record again.\n\n$info');
        // fall through to the real server
        return Response(418,
            body: jsonEncode({
              'message': 'no response found, fallthrough',
              'mode': serverMode.name,
              'request': ri.path,
              'scenario': s,
            }));
      }
      log.warning('The request ${ri.method} ${ri.uri} was not found in the '
          'scenario $s, which returns a 501 to the client. You may need to record again.\n\n${_buildRequestInfo(ri)}');
      return Response(501,
          headers: {
            'x-cache-replay': 'true',
          },
          body: jsonEncode({
            'message': 'no response found',
            'mode': serverMode.name,
            'request': ri.path,
            'scenario': s,
          }));
    }

    final headers = res.headers;
    headers['x-cache-replay'] = 'true';
    return Response(res.statusCode, headers: headers, body: res.body);
  }

  String _buildRequestInfo(RequestInfo ri) {
    final sb = StringBuffer();
    sb.writeln('Request: ${ri.method} ${ri.uri}');
    sb.writeln('Headers:');
    for (final key in ri.headers.keys) {
      sb.writeln('  $key: ${ri.headers[key]}');
    }
    if (ri.body != null && ri.body!.isNotEmpty) {
      sb.writeln('Body:');
      sb.writeln(ri.body);
    }
    return sb.toString();
  }
}

enum Mode {
  disabled,
  record,
  replay,
  replay_fallthrough,
}

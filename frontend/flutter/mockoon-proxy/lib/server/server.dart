import 'dart:convert';
import 'dart:io';

import 'package:mockoon_proxy/request_info.dart';
import 'package:mockoon_proxy/response_info.dart';
import 'package:shelf/shelf.dart';
import 'package:shelf_router/shelf_router.dart';
import 'package:shelf/shelf_io.dart' as io;
import 'package:shelf_cors_headers/shelf_cors_headers.dart';

import 'scenario_manager.dart';

Directory targetDirectory = Directory('scenarios');
ScenarioManager? manager;

ScenarioManager getManager({String? scenarioDir}) {
  if (manager == null) {
    // look for an env variable called SCENARIO_DIR
    // if it exists, use that as the directory to store the scenarios
    // otherwise, use the default
    scenarioDir =
        scenarioDir ?? Platform.environment['SCENARIO_DIR'] ?? 'scenarios';

    manager = ScenarioManager(scenarioDir);
  }
  return manager!;
}

void main() async {
  // Create a pipeline that handles requests.
  var handler = Pipeline()
      .addMiddleware(logRequests())
      .addMiddleware(corsHeaders())
      .addHandler(_router);

  getManager();
  // Create a server and bind it to a specific address and port.
  var server = await io.serve(handler, 'localhost', 8080);

  print('Serving at http://${server.address.host}:${server.port}');
}

final _router = Router()
  ..get('/scenarios', _listHandler)
  ..post('/scenarios/<scenario>/start', _startHandler)
  ..post('/scenarios/<scenario>/save', _saveHandler)
  ..get('/scenarios/<scenario>/close', _closeHandler)
  ..post('/scenarios/<scenario>/fetch', _replayHandler);

// starts a new scenario by creating the directory
Future<Response> _listHandler(Request request) async {
  final scenarios = await getManager().listScenarios();
  return Response.ok(jsonEncode({
    'scenarios': scenarios,
  }));
}

// starts a new scenario by creating the directory
Future<Response> _startHandler(Request request, String scenario) async {
  await getManager().startScenario(scenario);
  return Response.ok(jsonEncode({
    'message': 'created',
    'scenario': scenario,
  }));
}

// starts a new scenario by procesing all of the files into a
// mockoon file
Future<Response> _closeHandler(Request request, String scenario) async {
  await getManager().closeScenario(scenario);
  return Response.ok(jsonEncode({
    'message': 'closed',
    'scenario': scenario,
  }));
}

// saves a file into the directory
Future<Response> _saveHandler(Request request, String scenario) async {
  final body = await request.readAsString();

  final json = jsonDecode(body);

  final response = ResponseInfo.fromJson(json);
  await getManager().saveResponse(scenario, response);
  return Response.ok(jsonEncode({
    'message': 'saved',
    'request': response.request.path,
    'scenario': scenario,
  }));
}

Future<Response> _replayHandler(Request request, String scenario) async {
  final body = await request.readAsString();

  final json = jsonDecode(body);

  final ri = RequestInfo.fromJson(json);

  final response = await getManager().getResponse(scenario, ri);
  if (response == null) {
    return Response(501,
        headers: {
          'x-cache-replay': 'true',
        },
        body: jsonEncode({
          'message': 'no response found',
          'request': ri.path,
          'scenario': scenario,
        }));
  }

  final headers = response.headers;
  headers['x-cache-replay'] = 'true';
  return Response(response.statusCode, headers: headers, body: response.body);
}

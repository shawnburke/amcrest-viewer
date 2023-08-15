import 'dart:convert';
import 'dart:io';
import 'package:args/args.dart';

import 'package:mockoon_proxy/models/request_info.dart';
import 'package:mockoon_proxy/models/response_info.dart';
import 'package:shelf/shelf.dart';
import 'package:shelf_router/shelf_router.dart';
import 'package:shelf/shelf_io.dart' as io;
import 'package:shelf_cors_headers/shelf_cors_headers.dart';

import 'scenario_manager.dart';

Directory targetDirectory = Directory('scenarios');
ScenarioManager? manager;
String? currentScenario;
Mode serverMode = Mode.disabled;

const serverModeHeader = 'x-mockoon-proxy-mode';

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


const argPort = 'port';

void main(List<String> arguments) async {
  // Create a pipeline that handles requests.
  var handler = Pipeline()
      .addMiddleware(logRequests())
      .addMiddleware(corsHeaders())
      .addHandler(_router);

  final mgr = getManager();

  final parser = ArgParser()..addOption(argPort, abbr: 'p', defaultsTo: Platform.environment['PORT'] ?? '9099' );
  ArgResults argResults = parser.parse(arguments);
  final host = Platform.environment['HOST'] ?? 'localhost';
  final port = argResults[argPort];

  final ctrlC = ProcessSignal.sigint;

  ctrlC.watch().listen((signal) {
    print('Cleaning up...');
    mgr.dispose();
    exit(0);
  });

  // Create a server and bind it to a specific address and port.
  var server = await io.serve(handler, host, int.parse(port));

  print('Serving at http://${server.address.host}:${server.port}');
}

final _router = Router()
  ..get('/scenarios', _listHandler)
  ..get('/mode', _getModeHandler)
  ..post('/mode/<value>', _setModeHandler)
  ..put('/scenarios/current/<scenario>', _setCurrentHandler)
  ..post('/scenarios/<scenario>/start', _startHandler)
  ..post('/scenarios/<scenario>/save', _saveHandler)
  ..get('/scenarios/<scenario>/close', _closeHandler)
  ..post('/scenarios/<scenario>/fetch', _fetchHandler);

// starts a new scenario by creating the directory
Future<Response> _listHandler(Request request) async {
  final scenarios = await getManager().listScenarios();
  return Response.ok(jsonEncode({
    'scenarios': scenarios,
  }));
}

Future<Response> _getModeHandler(Request request) async {
  return Response.ok(jsonEncode({
    'message': 'current mode',
    'mode': serverMode.name,
    'current': currentScenario,
  }));
}

Future<Response> _setModeHandler(Request request, String value) async {
  final String? current = request.url.queryParameters['scenario'];
  if (current != null) {
    currentScenario = current;
  }

  for (final v in Mode.values) {
    if (v.name == value) {
      serverMode = v;
      break;
    }
  }

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

Future<Response> _setCurrentHandler(Request request, String scenario) async {
  print('setting current scenario to $scenario');
  currentScenario = scenario;
  return Response.ok(jsonEncode({
    'message': 'set current',
    'scenario': scenario,
  }));
}

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
  await getManager().closeScenario(s);
  return Response.ok(jsonEncode({
    'message': 'closed',
    'mode': serverMode.name,
    'scenario': s,
  }));
}

// saves a file into the directory
Future<Response> _saveHandler(Request request, String scenario) async {
  if (serverMode == Mode.disabled) {
    // server is disabled, so client should
    // return the call as normal
    return Response(418,
        body: jsonEncode({
          'message': 'server mode is disabled',
          'mode': serverMode.name,
        }));
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
  if (serverMode == Mode.disabled) {
    // server is disabled, so client should
    // call as normal through to the real server.
    return Response(418,
        body: jsonEncode({
          'message': 'server mode is disabled',
          'mode': serverMode.name,
        }));
  }

  final s = checkCurrent(scenario);
  if (s == null) {
    return noCurrentSet;
  }

  final body = await request.readAsString();

  final json = jsonDecode(body);

  final ri = RequestInfo.fromJson(json);

  final res = await getManager().getResponse(s, ri);
  if (res == null) {
    if (serverMode == Mode.replay_fallthrough) {
      // fall through to the real server
      return Response(418,
          body: jsonEncode({
            'message': 'no response found, fallthrough',
            'mode': serverMode.name,
            'request': ri.path,
            'scenario': s,
          }));
    }

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

enum Mode {
  disabled,
  record,
  replay,
  replay_fallthrough,
}

import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:logging/logging.dart';
import 'package:mockoon_proxy/mockoon/mockoon_environment.dart';

import '../models/request_info.dart';
import '../models/response_info.dart';

const Duration maxStartDelay = Duration(seconds: 10);

class MockoonManager {
  final log = Logger('MockoonManager');
  Process? _server;
  DateTime? _startTime;
  final List<_ScenarioFile> _files = [];
  final Map<String, int> _scenarioPorts = {};
  _ServerStatus _status = _ServerStatus.stopped;
  int mockoonPort = 0;
  bool verbose;
  late final String? tempDir;

  MockoonManager({this.tempDir, this.verbose = false}) {
    _startTime = DateTime.fromMillisecondsSinceEpoch(0);
    const oneSec = Duration(seconds: 1);
    Timer.periodic(oneSec, (Timer t) async {
      // if files have changed, shut down the server so the next
      // request starts it up again.
      if (await _checkFiles()) {
        shutdown();
      }
    });
  }

  void dispose() {
    shutdown();
  }

  Future<void> setScenarioPort(String scenario, int port) async {
    _scenarioPorts[scenario] = port;
    final sf = await findScenario(scenario);
    if (sf != null) {
      sf.staticPort = port;
      shutdown();
    }
  }

  Future<void> addScenario(File scenarioFile) async {
    for (final f in _files) {
      if (f.file.path == scenarioFile.path) {
        return;
      }
    }
    final scenarioName = _ScenarioFile.getScenarioName(scenarioFile);
    _files.add(_ScenarioFile(scenarioFile,
        tmpDir: tempDir, staticPort: _scenarioPorts[scenarioName]));
    shutdown();
  }

  Future<_ScenarioFile?> findScenario(String scenarioName) {
    for (final f in _files) {
      if (f.scenarioName == scenarioName) {
        return Future.value(f);
      }
    }
    return Future.value(null);
  }

  Future<ResponseInfo> makeRequest(String scenarioName, RequestInfo req) async {
    final scenarioFile = await findScenario(scenarioName);

    if (scenarioFile == null) {
      throw Exception('Scenario not found: $scenarioName');
    }
    try {
      await _runServer();
    } on ArgumentError {
      return ResponseInfo(
          req,
          503,
          {},
          jsonEncode({
            'message': 'Scenario not found: $scenarioName',
          }));
    }

    final running = await scenarioFile.isRunning;

    if (!running) {
      throw Exception('Scenario not running: $scenarioName');
    }
    final port = await scenarioFile.port;
    final headers = req.headers;

    // add a header so we can disambiguate body
    // changes
    final bodyHash = (req.body?.hashCode ?? true).toString();
    headers[headerBodyHash] = bodyHash;

    final dio = Dio(BaseOptions(
      baseUrl: 'http://${scenarioFile.host}:${port}/',
      connectTimeout:
          Duration(seconds: 30), // give server plenty of time to start up.
      validateStatus: (status) => true,
    ));

    //final p = '/' + MockoonEnvironment.normalizePath(req.uri.path);

    final resp = await dio.request(req.uri.toString(),
        data: req.body,
        //queryParameters: req.queryParameters,
        options: Options(
          method: req.method,
          headers: headers,
          receiveDataWhenStatusError: true,
          sendTimeout: Duration(seconds: 1),
        ));

    var ri = ResponseInfo.fromDio(resp);

    switch (ri.statusCode) {
      case 200:
        log.info(
            "Returning cached response: ${req.method} ${req.uri} (bodyhash=$bodyHash)");
        break;
      case 404:
        // we differentiate between an expected 404 and an unexpected 404 by looking
        // for the 'x-mockoon-served' header
        if (!ri.headers.containsKey('x-mockoon-served')) {
          log.warning(
              '[$scenarioName] missing recorded data for ${req.method} ${req.uri} (bodyhash=$bodyHash)');
          ri = ResponseInfo(ri.request, 501, ri.headers, ri.body);
        }
        break;
    }
    return ri;
  }

  Future<void> shutdown() async {
    if (_server != null) {
      log.info("Shutting down mockoon server...");
      _server!.kill();
      _status = _ServerStatus.stopped;
    }
    _server = null;
  }

  Future<bool> _checkFiles() async {
    if (_startTime == null) {
      return false;
    }
    for (final f in _files) {
      if (!f.file.existsSync()) {
        continue;
      }
      if (f.file.lastModifiedSync().isAfter(_startTime!)) {
        return true;
      }
    }
    return false;
  }

  Future<void> _runServer() async {
    if (_status != _ServerStatus.stopped) {
      return;
    }

    _server = await _run();
  }

  Future<Process?> _run() async {
    if (_status == _ServerStatus.starting) {
      return Future.value(_server!);
    }

    // Run the command and get the process result
    var args = ['start'];

    var scenarios = <String>[];
    for (final f in _files) {
      final updatedPath = await f.targetFile();
      if (f.staticPort != null) {
        continue;
      }
      args.add('-d');
      args.add(updatedPath.path);
      scenarios.add(f.scenarioName);
    }

    if (scenarios.isEmpty) {
      return null;
    }

    _status = _ServerStatus.starting;

    Process? processResult;

    try {
      log.info(
          'Starting mockoon server (Scenarios: ${scenarios.join(',')})...');
      processResult = await Process.start('mockoon-cli', args);
    } on ProcessException catch (e) {
      log.severe('Error starting mockoon-cli process: ${e.message}');
      if (e.message.contains('No such file or directory')) {
        log.shout(
            'Please install mockoon-cli (npm install -g @mockoon/cli), make sure it\'s on the path and try again.');
        exit(1);
      }
      _status = _ServerStatus.stopped;
      return Future.error(e);
    }
    _startTime = DateTime.now();

    processResult.exitCode.then((value) {
      _status = _ServerStatus.stopped;
      _startTime = null;
      log.info('Mockoon-cli server stopped with exit code $value');
    });

    final stdout = processResult.stdout.transform(utf8.decoder);
    final stderr = processResult.stderr.transform(utf8.decoder);

    stdout.listen((data) {
      log.fine('Mockoon-cli stdout: $data');
    });

    stderr.listen((data) {
      log.info('Mockoon-cli stderr: $data');
    });

    if (_status == _ServerStatus.stopped) {
      return processResult;
    }

    _status = _ServerStatus.running;

    return processResult;
  }
}

enum _ServerStatus { starting, running, stopped }

class _ScenarioFile {
  final File file;
  late final String scenarioName;
  final String host = 'localhost';
  int _port = 0;
  int? staticPort;
  final String? tmpDir;

  _ScenarioFile(this.file, {this.tmpDir, this.staticPort}) {
    scenarioName = getScenarioName(file);
  }

  Future<int> get port async {
    if (staticPort != null) {
      return staticPort!;
    }

    if (_port == 0) {
      final server = await ServerSocket.bind(InternetAddress.loopbackIPv4, 0);

      // Close the server when you're done
      _port = server.port;
      await server.close();
    }
    return _port;
  }

  String get targetFilePath {
    var targetPath = file.path.replaceAll(".json", ".tmp.json");

    if (tmpDir != null) {
      final tempFile = File('$tmpDir/${file.path}');
      targetPath = tempFile.path;
    }
    return targetPath;
  }

  void setStaticPort(int port) {
    staticPort = port;
  }

  Future<File> targetFile() async {
    if (staticPort != null) {
      return file;
    }
    // copy the file to a temp location and update the port
    // to our port
    final p = await port;

    final portRe = RegExp(r'"port":\s*(\d+)');
    try {
      final content = file.readAsStringSync();
      final updated = content.replaceAllMapped(portRe, (match) {
        return '"port": ${p}';
      });

      final tempFile = File(targetFilePath);
      if (!tempFile.parent.existsSync()) {
        tempFile.parent.createSync(recursive: true);
      }
      tempFile.writeAsStringSync(updated);
      return tempFile;
    } on PathNotFoundException {
      throw ArgumentError('Scenario file not found: ${file.path}');
    }
  }

  @override
  String toString() {
    return '$scenarioName: $port';
  }

  @override
  int get hashCode => file.hashCode;

  @override
  operator ==(Object other) {
    return other is _ScenarioFile && other.file == file;
  }

  Future<bool> get isRunning async {
    final port = await this.port;
    return await _waitForPort(host, port, maxStartDelay);
  }

  static Future<bool> _waitForPort(
      String host, int port, Duration maxWait) async {
    final maxWaitUntil = DateTime.now().add(maxWait);
    while (!await _checkPortStatus('localhost', port)) {
      if (DateTime.now().isAfter(maxWaitUntil)) {
        return false;
      }
      await Future.delayed(Duration(milliseconds: 100));
    }
    return true;
  }

  static Future<bool> _checkPortStatus(String host, int port) async {
    try {
      final socket = await Socket.connect(host, port);
      socket.close();
      return true; // Port is open
    } catch (e) {
      return false; // Port is closed or unreachable
    }
  }

  static String getScenarioName(File file) {
    return file.parent.path.split('/').last;
  }
}

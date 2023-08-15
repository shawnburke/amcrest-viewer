import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';

import '../models/request_info.dart';
import '../models/response_info.dart';

const Duration maxStartDelay = Duration(seconds: 3);

class MockoonManager {
  Process? _server;
  DateTime? _startTime;
  final List<_ScenarioFile> _files = [];
  _ServerStatus _status = _ServerStatus.stopped;
  int mockoonPort = 0;
  late final String _tempDir;

  MockoonManager() {
    _startTime = DateTime.fromMillisecondsSinceEpoch(0);
    final tmp = Directory.systemTemp.path;
    _tempDir = "$tmp/mockoon-proxy/${DateTime.now().millisecondsSinceEpoch}/";

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

  Future<void> addScenario(File scenarioFile) async {
    for (final f in _files) {
      if (f.file.path == scenarioFile.path) {
        return;
      }
    }
    _files.add(_ScenarioFile(scenarioFile));
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
      throw Exception('Scenario not running: $scenarioName');
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
    final dio = Dio(BaseOptions(
      baseUrl: 'http://${scenarioFile.host}:${port}/',
      headers: req.headers,
      connectTimeout: Duration(seconds: 3),
      validateStatus: (status) => true,
    ));

    final resp = await dio.request(req.uri.path,
        data: req.body,
        options: Options(
          method: req.method,
          receiveDataWhenStatusError: true,
          sendTimeout: Duration(seconds: 1),
        ));

    var ri = ResponseInfo.fromDio(resp);

    // we differentiate between an expected 404 and an unexpected 404 by looking
    // for the 'x-mockoon-served' header
    if (ri.statusCode == 404 && !ri.headers.containsKey('x-mockoon-served')) {
      ri = ResponseInfo(ri.request, 501, ri.headers, ri.body);
    }
    return ri;
  }

  Future<void> shutdown() async {
    if (_server != null) {
      print("Shutting down mockoon server...");
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

  Future<Process> _run() async {
    if (_status == _ServerStatus.starting) {
      return Future.value(_server!);
    }

    _status = _ServerStatus.starting;
    // Run the command and get the process result
    var args = ['start'];

    for (final f in _files) {
      final updatedPath = await f.targetFile(_tempDir);
      args.add('-d');
      args.add(updatedPath.path);
    }

    final processResult = await Process.start('mockoon-cli', args);
    _startTime = DateTime.now();

    processResult.exitCode.then((value) {
      _status = _ServerStatus.stopped;
      _startTime = null;
      print('Mockoon server stopped with exit code $value');
    });

    final stdout = processResult.stdout.transform(utf8.decoder);
    final stderr = processResult.stderr.transform(utf8.decoder);

    stdout.listen((data) {
      print('Mockoon stdout: $data');
    });

    stderr.listen((data) {
      print('Mockoon stderr: $data');
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

  _ScenarioFile(this.file) {
    scenarioName = file.parent.path.split('/').last;
  }

  Future<int> get port async {
    if (_port == 0) {
      final server = await ServerSocket.bind(InternetAddress.loopbackIPv4, 0);

      // Close the server when you're done
      _port = server.port;
      await server.close();
    }
    return _port;
  }

  Future<File> targetFile(String tempDir) async {
    // copy the file to a temp location and update the port
    // to our port
    final p = await port;

    final tempFile = File('$tempDir/${file.path}');

    if (!tempFile.parent.existsSync()) {
      tempFile.parent.createSync(recursive: true);
    }
    final portRe = RegExp(r'"port":\s*(\d+)');
    try {
      final content = file.readAsStringSync();
      final updated = content.replaceAllMapped(portRe, (match) {
        return '"port": ${p}';
      });

      tempFile.writeAsStringSync(updated);
      return tempFile;
    } on PathNotFoundException {
      throw ArgumentError('Scenario file not found: ${file.path}');
    }
  }

  @override
  String toString() {
    return '$_port: $scenarioName';
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
}

import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';

import 'package:dio/dio.dart';

import '../request_info.dart';
import '../response_info.dart';

class MockoonManager {
  Process? _server;
  DateTime? _startTime;
  final List<File> _files = [];
  _ServerStatus _status = _ServerStatus.stopped;
  int mockoonPort = 3000;
  late final int _randomKey;

  MockoonManager() {
    _startTime = DateTime.fromMillisecondsSinceEpoch(0);
    _randomKey = DateTime.now().millisecondsSinceEpoch;

    const oneSec = Duration(seconds: 1);
    Timer.periodic(oneSec, (Timer t) => _checkServer());
  }

  Future<void> addScenario(File scenarioFile) async {
    for (final f in _files) {
      if (f.path == scenarioFile.path) {
        return;
      }
    }
    _files.add(scenarioFile);
    await _checkServer(restart: true);
  }

  Future<String> _setScenarioPort(File file) {
    // copy the file to a temp location and update it's port
    // to be the mockoon port.

    final tempDir = Directory.systemTemp;
    final tempFile = File('${tempDir.path}/${_randomKey}/${file.path}');

    if (!tempFile.parent.existsSync()) {
      tempFile.parent.createSync(recursive: true);
    }
    final portRe = RegExp(r'"port":\s*(\d+)');
    final content = file.readAsStringSync();
    final updated = content.replaceAllMapped(portRe, (match) {
      return '"port": $mockoonPort';
    });
    tempFile.writeAsStringSync(updated);
    return Future.value(tempFile.path);
  }

  Future<File?> findScenario(String scenarioName) {
    for (final f in _files) {
      if (f.parent.path.endsWith(scenarioName)) {
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
    final dio = Dio(BaseOptions(
      baseUrl: 'http://localhost:$mockoonPort/' + scenarioName,
      headers: req.headers,
    ));

    final resp = await dio.request(req.uri.path,
        data: req.body,
        options: Options(
          method: req.method,
          receiveDataWhenStatusError: true,
        ));

    return ResponseInfo.fromDio(resp);
  }

  Future<void> shutdown() async {
    switch (_status) {
      case _ServerStatus.starting:
      case _ServerStatus.stopped:
        return;
      default:
        break;
    }

    if (_server != null) {
      print("Shutting down mockoon server...");
      _server!.kill();
      _status = _ServerStatus.stopped;
    }
    _server = null;
  }

  bool _shouldRestart() {
    if (_files.isEmpty) {
      return false;
    }

    if (_server == null) {
      return true;
    }

    if (_startTime == null) {
      return true;
    }

    for (final f in _files) {
      if (f.lastModifiedSync().isAfter(_startTime!)) {
        return true;
      }
    }

    return false;
  }

  Future<void> _checkServer({bool restart = false}) async {
    if (restart || _shouldRestart()) {
      await shutdown();
      _server = await run(_files.map((e) => e.path).toList());
    }
  }

  Future<Process> run(List<String> files) async {
    _status = _ServerStatus.starting;
    // Run the command and get the process result
    var args = ['start'];

    for (final f in files) {
      final updatedPath = await _setScenarioPort(File(f));
      args.add('-d');
      args.add(updatedPath);
    }

    final processResult = await Process.start('mockoon-cli', args);
    _status = _ServerStatus.running;
    processResult.exitCode.then((value) {
      _status = _ServerStatus.stopped;
      print('Mockoon server stopped with exit code $value');
    });
    _startTime = DateTime.now();

    final stdout = processResult.stdout.transform(utf8.decoder);
    final stderr = processResult.stderr.transform(utf8.decoder);

    stdout.listen((data) {
      print('Mockoon stdout: $data');
    });

    stderr.listen((data) {
      print('Mockoon stderr: $data');
    });

    if (_status != _ServerStatus.running) {
      return processResult;
    }

    await Future.delayed(Duration(seconds: 2), () {
      print('Mockoon server starting...');
    });
    return processResult;
  }
}

enum _ServerStatus { starting, running, stopped }

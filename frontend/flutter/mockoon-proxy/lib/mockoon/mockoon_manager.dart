import 'dart:async';
import 'dart:convert';
import 'dart:io';

class MockoonManager {
  Process? _server;
  DateTime? _startTime;
  final List<File> _files = [];

  MockoonManager() {
    _startTime = DateTime.fromMillisecondsSinceEpoch(0);

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

  Future<void> shutdown() async {
    if (_server != null) {
      print("Shutting down mockoon server...");
      _server!.kill();
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
    // Run the command and get the process result
    var args = ['start'];

    for (final f in files) {
      args.add('-d');
      args.add(f);
    }

    final processResult = await Process.start('mockoon-cli', args);
    _startTime = DateTime.now();

    final stdout = processResult.stdout.transform(utf8.decoder);
    final stderr = processResult.stderr.transform(utf8.decoder);

    stdout.listen((data) {
      print('Mockoon stdout: $data');
    });

    stderr.listen((data) {
      print('Mockoon stderr: $data');
    });

    await Future.delayed(Duration(seconds: 2), () {
      print('Mockoon server starting...');
    });
    return processResult;
  }
}

import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:logging/logging.dart';
import 'package:mockoon_proxy/mockoon/mockoon_manager.dart';
import 'package:mockoon_proxy/models/request_info.dart';
import 'package:mockoon_proxy/models/response_info.dart';

import '../mockoon/mockoon_environment.dart';
import '../mockoon/mockoon_model.dart';

const mockoonFile = 'mockoon.json';

class ScenarioManager {
  final log = Logger('ScenarioManager');
  late final Directory targetDirectory;
  int mockoonPort = 3333;
  late final MockoonManager manager;

  ScenarioManager(String directoryPath,
      {this.mockoonPort = 3000, bool useTempDir = false}) {
    targetDirectory = Directory(directoryPath);
    String? tempDir;
    if (useTempDir) {
      final tmp = Directory.systemTemp.path;
      tempDir = "$tmp/mockoon-proxy/${DateTime.now().millisecondsSinceEpoch}/";
    }
    manager = MockoonManager(tempDir: tempDir);
    _handleWriteJobs();
  }

  void dispose() {
    manager.dispose();
    writeController.close();
  }

  Future<void> _ensureDirectory() async {
    final exists = await targetDirectory.exists();
    if (!exists) {
      await targetDirectory.create(recursive: true);
    }
  }

  Future<void> setScenarioPort(String scenario, int port) async {
    await manager.setScenarioPort(scenario, port);
  }

  Future<List<String>> listScenarios({String? filter}) async {
    await _ensureDirectory();
    final dirs = await targetDirectory
        .list()
        .where((element) => element is Directory)
        .map((e) => e.path)
        .toList();

    final scenarioDirs = dirs
        .where((element) {
          if (filter == null) {
            return true;
          }
          final f = File('$element/${filter}');
          return f.existsSync();
        })
        .map((e) => e.split('/').last)
        .toList();

    return scenarioDirs;
  }

  Future<Directory> _getScenarioDirectory(String scenarioName) async {
    await _ensureDirectory();
    final pathName = scenarioName.replaceAll(' ', '_');
    final scenarioDirectory = Directory('${targetDirectory.path}/$pathName');

    final exists = await scenarioDirectory.exists();
    if (!exists) {
      await scenarioDirectory.create();
    }
    return scenarioDirectory;
  }

  Future<bool> startScenario(String scenarioName, {bool clear = false}) async {
    var reset = false;
    if (clear) {
      // delete the directory if it exists
      final scenarioDirectory = await _getScenarioDirectory(scenarioName);

      if (scenarioDirectory.existsSync()) {
        reset = true;
        await scenarioDirectory.delete(recursive: true);
      }
    }
    await _getScenarioDirectory(scenarioName);
    return reset;
  }

  String _getFileForRequest(RequestInfo ri) {
    var key = '${ri.uri}'
        .replaceAll('/', '_')
        .replaceAll(' ', '_')
        .replaceAll('=', '_')
        .replaceAll('?', '_')
        .replaceAll('&', '_');

    if (ri.body != null) {
      key += '_${ri.body.hashCode}';
    }

    return 'response-$key.json';
  }

  Future<void> saveResponse(
      String scenarioName, ResponseInfo responseInfo) async {
    final scenarioDirectory = await _getScenarioDirectory(scenarioName);

    final path = _getFileForRequest(responseInfo.request);

    final responseFile = File('${scenarioDirectory.path}/${path}');

    writeController.add(_WriteJob(responseFile, responseInfo));
  }

  Future<ResponseInfo?> getResponse(
      String scenarioName, RequestInfo req) async {
    return await _loadMockoon(scenarioName, req);
  }

  Future<void> _ensureServer(String scenarioName) async {
    final scenarioDir = await _getScenarioDirectory(scenarioName);

    final scenarioFile = File('${scenarioDir.path}/$mockoonFile');

    return manager.addScenario(scenarioFile);
  }

  Future<ResponseInfo> _loadMockoon(
      String scenarioName, RequestInfo req) async {
    await _ensureServer(scenarioName);
    return await manager.makeRequest(scenarioName, req);
  }

  //
  // File writing stuff
  //
  // We might get many instances of a route at about the same time
  // currently we do not set timestamps on the routes so the same file
  // might get written concurrently so the below serializes file writes.
  //

  final writeController = StreamController<_WriteJob>();

  Future<void> _handleWriteJobs() async {
    await for (final job in writeController.stream) {
      await _handleWriteJob(job);
    }
  }

  Future<void> _handleWriteJob(_WriteJob job) async {
    if (await job.file.exists()) {
      await job.file.delete();
    }
    final ri = job.ri.clean();
    var json = toJson(ri.toJson());

    if (await job.file.exists()) {
      log.fine('overwriting existing response file: ${job.file.path}');
    }
    await job.file.writeAsString(json, flush: true);
    log.info('Saved ${ri.request.uri} response to ${job.file.path}');
  }

  static MockoonModel fromDirectory(String name,
      {int port = 3000, String pathPrefix = '', Logger? log}) {
    final dir = Directory(name);

    final files = dir.listSync();

    final responses = files
        .map((e) {
          final filename = e.path.split('/').last;

          if (!filename.startsWith('response-')) {
            return null;
          }

          final contents = File(e.path).readAsStringSync();

          try {
            final ri = ResponseInfo.fromJson(jsonDecode(contents));
            return ri;
          } catch (e) {
            (log ?? Logger.root).severe('Error parsing $filename: $e');
            rethrow;
          }
        })
        .where((element) => element != null)
        .map((element) => element!)
        .toList();

    return MockoonEnvironmentBuilder.build(name, responses,
        port: port, pathPrefix: pathPrefix);
  }

  Future<String> closeScenario(String scenarioName) async {
    final dir = await _getScenarioDirectory(scenarioName);

    final mockoon = fromDirectory(dir.path, log: log);

    final f = File('${dir.path}/$mockoonFile');
    f.writeAsStringSync(toJson(mockoon.toJson()));
    log.fine('Wrote mockoon file to ${f.path}');
    return f.path;
  }

  static String toJson(dynamic json) {
    return JsonEncoder.withIndent(" ").convert(json);
  }
}

class _WriteJob {
  final File file;
  final ResponseInfo ri;

  _WriteJob(this.file, this.ri);
}

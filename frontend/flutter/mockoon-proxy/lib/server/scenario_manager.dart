import 'dart:convert';
import 'dart:io';

import 'package:mockoon_proxy/mockoon/mockoon_environment.dart';
import 'package:mockoon_proxy/request_info.dart';
import 'package:mockoon_proxy/response_info.dart';

class ScenarioManager {
  late final Directory targetDirectory;

  ScenarioManager(String directoryPath) {
    targetDirectory = Directory(directoryPath);
  }

  Future<void> _ensureDirectory() async {
    final exists = await targetDirectory.exists();
    if (!exists) {
      await targetDirectory.create();
    }
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

  Future<void> startScenario(String scenarioName, {bool clear = false}) async {
    if (clear) {
      // delete the directory if it exists
      final scenarioDirectory = await _getScenarioDirectory(scenarioName);
      await scenarioDirectory.delete(recursive: true);
    }
    await _getScenarioDirectory(scenarioName);
  }

  String _getFileForRequest(RequestInfo ri) {
    final key = '${ri.uri}'
        .replaceAll('/', '_')
        .replaceAll(' ', '_')
        .replaceAll('=', '_')
        .replaceAll('?', '_')
        .replaceAll('&', '_');

    return 'response-$key.json';
  }

  Future<void> saveResponse(
      String scenarioName, ResponseInfo responseInfo) async {
    final scenarioDirectory = await _getScenarioDirectory(scenarioName);

    final path = _getFileForRequest(responseInfo.request);

    final responseFile = File('${scenarioDirectory.path}/${path}');

    if (await responseFile.exists()) {
      print('WARN: overwriting existing response file: ${responseFile.path}');
    }

    final json = toJson(responseInfo.toJson());

    await responseFile.writeAsString(json);
    print('Saved ${responseInfo.request.uri} response to ${responseFile.path}');
  }

  Future<ResponseInfo?> getResponse(
      String scenarioName, RequestInfo req) async {
    final scenarioDirectory = await _getScenarioDirectory(scenarioName);

    final path = _getFileForRequest(req);
    final responseFile = File('${scenarioDirectory.path}/${path}');

    if (!await responseFile.exists()) {
      return null;
    }

    final json = await responseFile.readAsString();

    final responseInfo = ResponseInfo.fromJson(jsonDecode(json));

    return responseInfo;
  }

  Future<void> closeScenario(String scenarioName) async {
    final dir = await _getScenarioDirectory(scenarioName);

    final mockoon = MockoonEnvironment.fromDirectory(dir.path);

    final f = File('${dir.path}/mockoon.json');
    f.writeAsStringSync(toJson(mockoon.toJson()));
    print('Wrote mockoon to ${f.path}');
  }

  static String toJson(dynamic json) {
    return JsonEncoder.withIndent(" ").convert(json);
  }
}

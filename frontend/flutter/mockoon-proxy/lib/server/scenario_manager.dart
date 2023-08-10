import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:mockoon_proxy/mockoon/mockoon_environment.dart';
import 'package:mockoon_proxy/request_info.dart';
import 'package:mockoon_proxy/response_info.dart';
import 'package:mockoon_proxy/server/server.dart';

class ScenarioManager {
  late final Directory targetDirectory;
  int mockoonPort = 3333;

  ScenarioManager(String directoryPath, {this.mockoonPort = 3333}) {
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
    return await _loadMockoon(scenarioName, req);
    // return _loadFile(scenarioName, req);
  }

  Future<ResponseInfo> _loadMockoon(
      String scenarioName, RequestInfo req) async {
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

  Future<ResponseInfo?> _loadFile(String scenarioName, RequestInfo req) async {
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

  Process? mockoonProcess;

  Future<void> reload() async {
    if (mockoonProcess != null) {
      await mockoonProcess!.kill();
    }

    final scenarios = await getManager().listScenarios(filter: 'mockoon.json');
    mockoonProcess =
        await runMockoon(scenarios.map((e) => e + '/mockoon.json').toList());
  }

  Future<Process> launchServer(List<String> dataFiles) async {
    // Run the command and get the process result
    var args = ['start'];

    for (final f in dataFiles) {
      args.add('-d');
      args.add(f);
    }

    return Process.start('mockoon-cli', args);
  }

  Future<Process> runMockoon(List<String> files) async {
    final processResult = await launchServer(files);

    final stdout = processResult.stdout.transform(utf8.decoder);
    final stderr = processResult.stderr.transform(utf8.decoder);

    stdout.listen((data) {
      print('Mockoon stdout: $data');
    });

    stderr.listen((data) {
      print('Mockoon stderr: $data');
    });
    return processResult;
  }

  static String toJson(dynamic json) {
    return JsonEncoder.withIndent(" ").convert(json);
  }
}

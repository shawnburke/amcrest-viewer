import 'dart:convert';
import 'dart:io';

import 'package:mockoon_proxy/mockoon/mockoon_environment.dart';
import 'package:mockoon_proxy/server/scenario_manager.dart';
import 'package:test/test.dart';

void main() {
  test('loads files', () {
    print(Directory.current.path);

    final mockoon = MockoonEnvironment.fromDirectory('test/data/test_scenario');

    final json = JsonEncoder.withIndent(" ").convert(mockoon.toJson());

    final f = File('test_scenario_output.json');
    f.writeAsStringSync(json);
    print('Wrote to ${f.path}');
  });

  test('list scenarios', () async {
    final manager = ScenarioManager('test/data');
    final scenarios = await manager.listScenarios();

    expect(scenarios.first, 'test_scenario');
  });
}

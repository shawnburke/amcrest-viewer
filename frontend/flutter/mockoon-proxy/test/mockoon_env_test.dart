import 'dart:convert';
import 'dart:io';

import 'package:mockoon_proxy/mockoon/mockoon_environment.dart';
import 'package:mockoon_proxy/response_info.dart';
import 'package:test/test.dart';

void main() {
  test('loads files', () {
    print(Directory.current.path);
    final dir = Directory('test/data/test_scenario');

    final files = dir.listSync();

    final responses = files.map((e) {
      final contents = File(e.path).readAsStringSync();

      final ri = ResponseInfo.fromJson(jsonDecode(contents));
      return ri;
    });

    final mockoon = MockoonEnvironment.build(
        'test_scenario', responses.toList(),
        port: 3333);

    expect(mockoon.routes!.length, 3);

    final json = JsonEncoder.withIndent(" ").convert(mockoon.toJson());

    final f = File('test_scenario_output.json');
    f.writeAsStringSync(json);
    print('Wrote to ${f.path}');
  });
}

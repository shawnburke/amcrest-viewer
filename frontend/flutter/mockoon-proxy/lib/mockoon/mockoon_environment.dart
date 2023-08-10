import 'dart:convert';
import 'dart:io';

import 'package:mockoon_proxy/mockoon/mockoon_model.dart';
import 'package:uuid_type/uuid_type.dart';

import '../response_info.dart';

class MockoonEnvironment {
  static Uuid getForName(String name) {
    return NameUuidGenerator(NameUuidGenerator.urlNamespace)
        .generateFromString("mockoon://" + name);
  }

  static MockoonModel fromDirectory(String name,
      {int port = 3000, String pathPrefix = ''}) {
    final dir = Directory(name);

    final files = dir.listSync();

    final responses = files.map((e) {
      final contents = File(e.path).readAsStringSync();

      final ri = ResponseInfo.fromJson(jsonDecode(contents));
      return ri;
    });

    return build(name, responses.toList(), port: port, pathPrefix: pathPrefix);
  }

  static MockoonModel build(String name, List<ResponseInfo> responses,
      {int port = 3000, String pathPrefix = ''}) {
    final uuid = getForName(name);

    final routes = <Routes>[];
    final model = MockoonModel(
      uuid: uuid.toString(),
      name: 'Generated Scenario: $name',
      port: port,
      routes: routes,
      rootChildren: [],
      cors: true,
      lastMigration: 28,
      latency: 0,
      hostname: "",
      folders: [],
      headers: [],
      data: [],
      proxyMode: false,
      proxyHost: "",
      proxyRemovePrefix: false,
      tlsOptions: TlsOptions(
        enabled: false,
        type: "CERT",
        pfxPath: "",
        certPath: "",
        keyPath: "",
        passphrase: "",
      ),
    );

    final responseMap = Map<String, List<ResponseInfo>>();

    for (final r in responses) {
      final key = r.request.method + '+' + r.request.uri.path;
      final list = responseMap[key] ?? [];
      list.add(r);
      responseMap[key] = list;
    }

    for (final g in responseMap.entries) {
      final uri = g.value.first.request.uri;

      var path = uri.path;

      if (pathPrefix.isNotEmpty) {
        path = pathPrefix + path;
      } else {
        path = path.substring(1);
      }

      final route = Routes(
        uuid: getForName(g.key).toString(),
        type: 'http',
        method: g.value.first.request.method,
        endpoint: path,
        responses: [],
      );
      routes.add(route);
      model.rootChildren ??= [];
      model.rootChildren!.add(RootChildren(
        uuid: route.uuid,
        type: 'route',
      ));

      for (final r in g.value) {
        final response = Responses(
          uuid: getForName(r.request.uri.toString()).toString(),
          statusCode: r.statusCode,
          headers: r.headers.entries
              .map((e) => Headers(key: e.key, value: e.value))
              .toList(),
          body: r.body,
          label: 'OK',
          latency: 0,
          bodyType: 'INLINE',
          sendFileAsBody: false,
          rules: r.request.queryParameters?.entries
              .map((e) => Rules(
                    target: 'query',
                    modifier: e.key,
                    operator: 'equals',
                    invert: false,
                    value: e.value?.toString() ?? '',
                  ))
              .toList(),
          rulesOperator: 'AND',
          disableTemplating: false,
          fallbackTo404: false,
          isDefault: false,
        );
        route.responses ??= [];
        response.headers!.add(Headers(key: 'x-mockoon-served', value: 'true'));
        route.responses!.add(response);
      }
    }
    return model;
  }
}

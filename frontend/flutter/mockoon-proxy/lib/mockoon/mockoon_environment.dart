import 'package:logging/logging.dart';
import 'package:mockoon_proxy/mockoon/mockoon_model.dart';
import 'package:uuid_type/uuid_type.dart';

import '../models/response_info.dart';

const headerBodyHash = 'x-mockoon-bodyhash';

class MockoonEnvironmentBuilder {
  static Uuid getForName(String name) {
    return NameUuidGenerator(NameUuidGenerator.urlNamespace)
        .generateFromString("mockoon://" + name);
  }

  static MockoonModel build(String name, List<ResponseInfo> responses,
      {int port = 3000, String pathPrefix = ''}) {
    final log = Logger('MockoonEnvironmentBuilder');

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
      endpointPrefix: "",
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

      var path = normalizePath(uri.path);

      final route = Routes(
        uuid: getForName(g.key).toString(),
        type: 'http',
        method: g.value.first.request.method.toLowerCase(),
        endpoint: path,
        responses: [],
      );
      routes.add(route);
      log.info(
          'Added route ${route.method!.toUpperCase()} /${route.endpoint} to $name');
      model.rootChildren ??= [];
      model.rootChildren!.add(RootChildren(
        uuid: route.uuid,
        type: 'route',
      ));

      var hasDefault = false;

      for (final r in g.value) {
        bool isDefault = (r.request.body?.isEmpty ?? true) &&
            (r.request.queryParameters?.isEmpty ?? true);

        if (isDefault) {
          hasDefault = true;
        }

        final bodyHash = r.request.body?.hashCode ?? 0;
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
          rules: r.request.uri.queryParameters.entries
              .map((e) => Rules(
                    target: 'query',
                    modifier: e.key,
                    operator: 'equals',
                    invert: false,
                    value: encodeQueryValue(e.value),
                  ))
              .toList(),
          rulesOperator: 'AND',
          disableTemplating: false,
          fallbackTo404: false,
          isDefault: isDefault,
        );
        response.headers!.add(Headers(key: 'x-mockoon-served', value: 'true'));

        // we add a rule to make sure the body matches.
        response.rules!.add(Rules(
          target: 'header',
          modifier: headerBodyHash,
          operator: 'equals',
          invert: false,
          value: r.body.isEmpty ? '0' : bodyHash.toString(),
        ));

        route.responses ??= [];
        route.responses!.add(response);
      }
      if (!hasDefault) {
        // add a dummy default route to ensure rules work properly.
        final defaultResponse = Responses(
          uuid:
              getForName('default-${route.method}${route.endpoint}').toString(),
          statusCode: 404,
          //headers: [Headers(key: 'x-mockoon-served', value: 'true')],
          body: '',
          label: 'Not Found',
          latency: 0,
          bodyType: 'INLINE',
          sendFileAsBody: false,
          rules: [],
          rulesOperator: 'AND',
          disableTemplating: false,
          fallbackTo404: false,
          isDefault: true,
        );
        route.responses!.add(defaultResponse);
      }
    }
    return model;
  }

  static String encodeQueryValue(dynamic val) {
    if (val == null) {
      return '';
    }
    return val.toString();
  }

  static String normalizePath(String p) {
    p = p.replaceAll('//', '/');
    if (p.startsWith('/')) {
      p = p.substring(1);
    }
    return p;
  }
}

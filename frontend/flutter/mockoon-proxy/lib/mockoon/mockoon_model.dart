class MockoonModel {
  String? uuid;
  int? lastMigration;
  String? name;
  String? endpointPrefix;
  int? latency;
  int? port;
  String? hostname;
  List<String>? folders;
  List<Routes>? routes;
  List<RootChildren>? rootChildren;
  bool? proxyMode;
  String? proxyHost;
  bool? proxyRemovePrefix;
  TlsOptions? tlsOptions;
  bool? cors;
  List<String>? headers;
  List<String>? data;

  MockoonModel(
      {this.uuid,
      this.lastMigration,
      this.name,
      this.endpointPrefix,
      this.latency,
      this.port,
      this.hostname,
      this.folders,
      this.routes,
      this.rootChildren,
      this.proxyMode,
      this.proxyHost,
      this.proxyRemovePrefix,
      this.tlsOptions,
      this.cors,
      this.headers,
      this.data});

  MockoonModel.fromJson(Map<String, dynamic> json) {
    uuid = json['uuid'];
    lastMigration = json['lastMigration'];
    name = json['name'];
    endpointPrefix = json['endpointPrefix'];
    latency = json['latency'];
    port = json['port'];
    hostname = json['hostname'];
    if (json['folders'] != null) {
      folders = <String>[];
      json['folders'].forEach((v) {
        folders!.add(v);
      });
    }
    if (json['routes'] != null) {
      routes = <Routes>[];
      json['routes'].forEach((v) {
        routes!.add(new Routes.fromJson(v));
      });
    }
    if (json['rootChildren'] != null) {
      rootChildren = <RootChildren>[];
      json['rootChildren'].forEach((v) {
        rootChildren!.add(new RootChildren.fromJson(v));
      });
    }
    proxyMode = json['proxyMode'];
    proxyHost = json['proxyHost'];
    proxyRemovePrefix = json['proxyRemovePrefix'];
    tlsOptions = json['tlsOptions'] != null
        ? new TlsOptions.fromJson(json['tlsOptions'])
        : null;
    cors = json['cors'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['uuid'] = this.uuid;
    data['lastMigration'] = this.lastMigration;
    data['name'] = this.name;
    data['endpointPrefix'] = this.endpointPrefix;
    data['latency'] = this.latency;
    data['port'] = this.port;
    data['hostname'] = this.hostname;
    if (this.folders != null) {
      data['folders'] = this.folders!.map((v) => v).toList();
    }
    if (this.routes != null) {
      data['routes'] = this.routes!.map((v) => v).toList();
    }
    if (this.rootChildren != null) {
      data['rootChildren'] = this.rootChildren!.map((v) => v.toJson()).toList();
    }
    data['proxyMode'] = this.proxyMode;
    data['proxyHost'] = this.proxyHost;
    data['proxyRemovePrefix'] = this.proxyRemovePrefix;
    if (this.tlsOptions != null) {
      data['tlsOptions'] = this.tlsOptions!.toJson();
    }
    data['cors'] = this.cors;
    if (this.headers != null) {
      data['headers'] = this.headers!.map((v) => v).toList();
    }

    if (this.data != null) {
      data['data'] = this.data!.map((v) => v).toList();
    }
    return data;
  }
}

class Routes {
  String? uuid;
  String? type;
  String? documentation;
  String? method;
  String? endpoint;
  List<Responses>? responses;
  bool? enabled;
  String? responseMode;

  Routes(
      {this.uuid,
      this.type,
      this.documentation,
      this.method,
      this.endpoint,
      this.responses,
      this.enabled,
      this.responseMode});

  Routes.fromJson(Map<String, dynamic> json) {
    uuid = json['uuid'];
    type = json['type'];
    documentation = json['documentation'];
    method = json['method'];
    endpoint = json['endpoint'];
    if (json['responses'] != null) {
      responses = <Responses>[];
      json['responses'].forEach((v) {
        responses!.add(new Responses.fromJson(v));
      });
    }
    enabled = json['enabled'];
    responseMode = json['responseMode'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['uuid'] = this.uuid;
    data['type'] = this.type;
    data['documentation'] = this.documentation;
    data['method'] = this.method;
    data['endpoint'] = this.endpoint;
    if (this.responses != null) {
      data['responses'] = this.responses!.map((v) => v.toJson()).toList();
    }
    data['enabled'] = this.enabled;
    data['responseMode'] = this.responseMode;
    return data;
  }
}

class Responses {
  String? uuid;
  String? body;
  int? latency;
  int? statusCode;
  String? label;
  List<Headers>? headers;
  String? bodyType;
  String? filePath;
  String? databucketID;
  bool? sendFileAsBody;
  List<Rules>? rules;
  String? rulesOperator;
  bool? disableTemplating;
  bool? fallbackTo404;
  bool? isDefault;
  String? crudKey;

  Responses(
      {this.uuid,
      this.body,
      this.latency,
      this.statusCode,
      this.label,
      this.headers,
      this.bodyType,
      this.filePath,
      this.databucketID,
      this.sendFileAsBody,
      this.rules,
      this.rulesOperator,
      this.disableTemplating,
      this.fallbackTo404,
      this.isDefault,
      this.crudKey});

  Responses.fromJson(Map<String, dynamic> json) {
    uuid = json['uuid'];
    body = json['body'];
    latency = json['latency'];
    statusCode = json['statusCode'];
    label = json['label'];
    if (json['headers'] != null) {
      headers = <Headers>[];
      json['headers'].forEach((v) {
        headers!.add(new Headers.fromJson(v));
      });
    }
    bodyType = json['bodyType'];
    filePath = json['filePath'];
    databucketID = json['databucketID'];
    sendFileAsBody = json['sendFileAsBody'];
    if (json['rules'] != null) {
      rules = <Rules>[];
      json['rules'].forEach((v) {
        rules!.add(new Rules.fromJson(v));
      });
    }
    rulesOperator = json['rulesOperator'];
    disableTemplating = json['disableTemplating'];
    fallbackTo404 = json['fallbackTo404'];
    isDefault = json['default'];
    crudKey = json['crudKey'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['uuid'] = this.uuid;
    data['body'] = this.body;
    data['latency'] = this.latency;
    data['statusCode'] = this.statusCode;
    data['label'] = this.label;
    if (this.headers != null) {
      data['headers'] = this.headers!.map((v) => v.toJson()).toList();
    }
    data['bodyType'] = this.bodyType;
    data['filePath'] = this.filePath;
    data['databucketID'] = this.databucketID;
    data['sendFileAsBody'] = this.sendFileAsBody;
    if (this.rules != null) {
      data['rules'] = this.rules!.map((v) => v.toJson()).toList();
    }
    data['rulesOperator'] = this.rulesOperator;
    data['disableTemplating'] = this.disableTemplating;
    data['fallbackTo404'] = this.fallbackTo404;
    data['default'] = this.isDefault;
    data['crudKey'] = this.crudKey;
    return data;
  }
}

class Headers {
  String? key;
  String? value;

  Headers({this.key, this.value});

  Headers.fromJson(Map<String, dynamic> json) {
    key = json['key'];
    value = json['value'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['key'] = this.key;
    data['value'] = this.value;
    return data;
  }
}

class Rules {
  String? target;
  String? modifier;
  String? value;
  bool? invert;
  String? operator;

  Rules({this.target, this.modifier, this.value, this.invert, this.operator});

  Rules.fromJson(Map<String, dynamic> json) {
    target = json['target'];
    modifier = json['modifier'];
    value = json['value'];
    invert = json['invert'];
    operator = json['operator'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['target'] = this.target;
    data['modifier'] = this.modifier;
    data['value'] = this.value;
    data['invert'] = this.invert;
    data['operator'] = this.operator;
    return data;
  }
}

class RootChildren {
  String? type;
  String? uuid;

  RootChildren({this.type, this.uuid});

  RootChildren.fromJson(Map<String, dynamic> json) {
    type = json['type'];
    uuid = json['uuid'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['type'] = this.type;
    data['uuid'] = this.uuid;
    return data;
  }
}

class TlsOptions {
  bool? enabled;
  String? type;
  String? pfxPath;
  String? certPath;
  String? keyPath;
  String? caPath;
  String? passphrase;

  TlsOptions(
      {this.enabled,
      this.type,
      this.pfxPath,
      this.certPath,
      this.keyPath,
      this.caPath,
      this.passphrase});

  TlsOptions.fromJson(Map<String, dynamic> json) {
    enabled = json['enabled'];
    type = json['type'];
    pfxPath = json['pfxPath'];
    certPath = json['certPath'];
    keyPath = json['keyPath'];
    caPath = json['caPath'];
    passphrase = json['passphrase'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['enabled'] = this.enabled;
    data['type'] = this.type;
    data['pfxPath'] = this.pfxPath;
    data['certPath'] = this.certPath;
    data['keyPath'] = this.keyPath;
    data['caPath'] = this.caPath;
    data['passphrase'] = this.passphrase;
    return data;
  }
}

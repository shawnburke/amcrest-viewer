import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:json_annotation/json_annotation.dart';

import 'response_info.dart';

part 'request_info.g.dart';

@JsonSerializable()
class RequestInfo {
  final String method;
  final Map<String, String> headers;
  final String path;

  final Map<String, dynamic>? queryParameters;
  final String? body;

  RequestInfo(
      this.method, this.path, this.headers, this.body, this.queryParameters) {}

  Uri get uri {
    return Uri(
        path: path,
        query: queryParameters?.entries
            .map((e) =>
                '${e.key}=${Uri.encodeQueryComponent(e.value.toString())}')
            .join('&'));
  }

  String get key => '${method}+${uri}';

  RequestInfo clean() {
    var b = this.body;
    if (this.body != null) {
      b = ResponseInfo.cleanData(b!);
    }


    final auth = headers['Authorization'];
    if (auth != null) {
      // invalidate the JWT
      headers['Authorization'] = auth.replaceAll(RegExp(r'\d'), 'x');
    }
    
    headers.remove('expiry');

    return RequestInfo(method, path, headers, b, queryParameters);
  }

  RequestInfo.fromDio(RequestOptions req)
      : this(req.method, _getPath(req), toStringStringMap(req.headers),
            stringify(req.data), _getQueryParams(req));

  static String _getPath(RequestOptions req) {
    return Uri.parse(req.path).path;
  }

  static Map<String, dynamic>? _getQueryParams(RequestOptions req) {
    final args = <String, dynamic>{};

    final uri = Uri.parse(req.path);

    args.addAll(uri.queryParameters);
    args.addAll(req.queryParameters);
    return args;
  }

  static String? stringify(dynamic data) {
    if (data == null) {
      return null;
    }
    if (data is String) {
      return data;
    }
    return jsonEncode(data);
  }

  static Map<String, String> toStringStringMap(Map<String, dynamic> map) {
    return map.map((key, value) => MapEntry(key, value.toString()));
  }

  factory RequestInfo.fromJson(Map<String, dynamic> json) =>
      _$RequestInfoFromJson(json);
  Map<String, dynamic> toJson() => _$RequestInfoToJson(this);
}

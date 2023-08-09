import 'package:dio/dio.dart';
import 'package:json_annotation/json_annotation.dart';

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

  RequestInfo.fromDio(RequestOptions req)
      : this(req.method, req.path, toStringStringMap(req.headers), req.data,
            req.queryParameters);

  static Map<String, String> toStringStringMap(Map<String, dynamic> map) {
    return map.map((key, value) => MapEntry(key, value.toString()));
  }

  factory RequestInfo.fromJson(Map<String, dynamic> json) =>
      _$RequestInfoFromJson(json);
  Map<String, dynamic> toJson() => _$RequestInfoToJson(this);
}

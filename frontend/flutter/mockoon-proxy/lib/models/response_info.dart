import 'package:dio/dio.dart';
import 'package:json_annotation/json_annotation.dart';
import 'request_info.dart';
import 'dart:convert';

part 'response_info.g.dart';

@JsonSerializable()
class ResponseInfo {
  final RequestInfo request;
  final int statusCode;
  final Map<String, String> headers;
  final String body;

  ResponseInfo(this.request, this.statusCode, this.headers, this.body);

  ResponseInfo clean() {
    final req = request.clean();
    final body = cleanData(this.body);
    // this will result in a "incompatible headers" error
    headers.remove('transfer-encoding');
    headers.remove('content-encoding');
    
    return ResponseInfo(req, statusCode, headers, body);
  }

  static String cleanData(String input) {
    final regex = RegExp(r'[\d,]+\.\d+');
    return input.replaceAll(regex, '123.45');
  }

  ResponseInfo.fromDio(Response res)
      : this(RequestInfo.fromDio(res.requestOptions), res.statusCode ?? 0,
            toMap(res.headers), jsonEncode(res.data));

  static Map<String, String> toMap(Headers h) {
    return h.map.map((key, value) => MapEntry(key, value.join(',')));
  }

  factory ResponseInfo.fromJson(Map<String, dynamic> json) =>
      _$ResponseInfoFromJson(json);
  Map<String, dynamic> toJson() => _$ResponseInfoToJson(this);
}

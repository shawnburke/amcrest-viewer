// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'request_info.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

RequestInfo _$RequestInfoFromJson(Map<String, dynamic> json) => RequestInfo(
      json['method'] as String,
      json['path'] as String,
      Map<String, String>.from(json['headers'] as Map),
      json['body'] as String?,
      json['queryParameters'] as Map<String, dynamic>?,
    );

Map<String, dynamic> _$RequestInfoToJson(RequestInfo instance) =>
    <String, dynamic>{
      'method': instance.method,
      'headers': instance.headers,
      'path': instance.path,
      'queryParameters': instance.queryParameters,
      'body': instance.body,
    };

// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'response_info.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ResponseInfo _$ResponseInfoFromJson(Map<String, dynamic> json) => ResponseInfo(
      RequestInfo.fromJson(json['request'] as Map<String, dynamic>),
      json['statusCode'] as int,
      Map<String, String>.from(json['headers'] as Map),
      json['body'] as String,
    );

Map<String, dynamic> _$ResponseInfoToJson(ResponseInfo instance) =>
    <String, dynamic>{
      'request': instance.request,
      'statusCode': instance.statusCode,
      'headers': instance.headers,
      'body': instance.body,
    };

import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:get_it/get_it.dart';
import 'package:openapi/openapi.dart';

import 'config.dart';
import 'interceptor.dart';

final GetIt locator = GetIt.instance;
final Config globalConfig = Config();
void setupLocator() {
  locator
    ..registerSingleton<Config>(globalConfig)
    ..registerSingleton<CameraApi>(createCameraApi(globalConfig))
    ..registerLazySingleton<CamViewerRepo>(() {
      final api = locator<CameraApi>();
      final defaultApi = api.getDefaultApi();
      var basePath = globalConfig.baseUri.toString();
      if (basePath.endsWith('/')) {
        basePath = basePath.substring(0, basePath.length - 1);
      }
      return CamViewerRepoImpl(defaultApi, basePath: basePath);
    });
}

CameraApi createCameraApi(Config config) {
  return CameraApi(basePath: config.baseUri.toString());
}

class CameraApi extends Openapi {
  final String basePath;

  CameraApi({required this.basePath})
      : super(basePathOverride: basePath, interceptors: [TrafficInterceptor()]);
}

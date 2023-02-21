import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:get_it/get_it.dart';

import 'config.dart';

final GetIt locator = GetIt.instance;
void setupLocator() {
  locator.registerFactory<CamViewerRepo>(() => CamViewerRepoImpl(url: baseURL));
}

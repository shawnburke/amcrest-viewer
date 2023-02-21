import 'package:flutter/foundation.dart';
import 'package:openapi/api.dart';

abstract class CamViewerRepo {
  Future<List<Camera>> getCameras({bool? includeLastSnapshot});
}

class CamViewerRepoImpl extends CamViewerRepo {
  late ApiClient client;

  CamViewerRepoImpl({required String url}) {
    client = ApiClient(basePath: url);
  }

  @override
  Future<List<Camera>> getCameras({bool? includeLastSnapshot}) async {
    var api = DefaultApi(client);
    var cams = await api.getCameras(latestSnapshot: includeLastSnapshot);

    if (cams == null) {
      return List<Camera>.empty();
    }

    return cams;
  }
}

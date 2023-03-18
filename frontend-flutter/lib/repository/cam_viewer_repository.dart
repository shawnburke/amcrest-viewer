import 'package:openapi/api.dart';

abstract class CamViewerRepo {
  Future<List<Camera>> getCameras({bool? includeLastSnapshot});
  Future<Camera?> getCamera(int id);
}

class CamViewerRepoImpl extends CamViewerRepo {
  late ApiClient client;
  late DefaultApi api;

  CamViewerRepoImpl({required String url}) {
    client = ApiClient(basePath: url);
    api = DefaultApi(client);
  }

  @override
  Future<List<Camera>> getCameras({bool? includeLastSnapshot}) async {
    var cams = await api.getCameras(latestSnapshot: includeLastSnapshot);

    if (cams == null) {
      return List<Camera>.empty();
    }

    return cams;
  }

  @override
  Future<Camera?> getCamera(int id) async {
    return api.getCamera(id);
  }
}

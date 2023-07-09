import 'package:openapi/api.dart';

abstract class CamViewerRepo {
  Future<List<Camera>> getCameras({bool? includeLastSnapshot});
  Future<Camera?> getCamera(int id);
  Future<List<CameraFile>> getFiles(int id, DateTime start, DateTime end);

  Future<String?> getLiveStreamURL(int id);
}

class CamViewerRepoImpl extends CamViewerRepo {
  late ApiClient client;
  late DefaultApi api;

  CamViewerRepoImpl({required String url}) {
    if (url.endsWith('/')) {
      url = url.substring(0, url.length - 1);
    }
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

  @override
  Future<List<CameraFile>> getFiles(
      int id, DateTime start, DateTime end) async {
    var result =
        await api.getCameraFiles(id.toString(), start: start, end: end);

    if (result == null) {
      return List<CameraFile>.empty();
    }
    for (var file in result) {
      if (!file.path.startsWith('http')) {
        file.path = client.basePath + file.path;
      }
    }
    return result;
  }

  @override
  Future<String?> getLiveStreamURL(int id) async {
    var result = await api.getCameraLiveStream(id, redirect: true);
    if (result == null) {
      return null;
    }
    return result.uri;
  }
}

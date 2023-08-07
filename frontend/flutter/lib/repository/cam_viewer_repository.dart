import 'package:openapi/openapi.dart';
import 'package:path/path.dart' as path;

abstract class CamViewerRepo {
  Future<List<Camera>> getCameras({bool? includeLastSnapshot});
  Future<Camera?> getCamera(int id);
  Future<List<CameraFile>> getFiles(int id, DateTime start, DateTime end);

  Future<String?> getLiveStreamURL(int id);
}

class CamViewerRepoImpl extends CamViewerRepo {
  final DefaultApi api;
  final String basePath;

  CamViewerRepoImpl(this.api, {String? basePath}) : basePath = basePath ?? '/';

  @override
  Future<List<Camera>> getCameras({bool? includeLastSnapshot}) async {
    var camsResponse =
        await api.getCameras(latestSnapshot: includeLastSnapshot);
    var cams = camsResponse.data;
    if (cams == null) {
      return List<Camera>.empty();
    }

    return cams.toList();
  }

  @override
  Future<Camera?> getCamera(int id) async {
    final response = await api.getCamera(id: id);
    return response.data;
  }

  @override
  Future<List<CameraFile>> getFiles(
      int id, DateTime start, DateTime end) async {
    var response = await api.getCameraFiles(
        id: id.toString(), start: start.toUtc(), end: end.toUtc());
    final result = response.data;
    if (result == null) {
      return List<CameraFile>.empty();
    }
    final resultList = result.toList();
    for (var i = 0; i < resultList.length; i++) {
      var file = resultList[i];
      if (!file.path.startsWith('http')) {
        final builder = file.toBuilder();
        builder.path = path.join(basePath, file.path);
        resultList[i] = builder.build();
      }
    }
    return resultList;
  }

  @override
  Future<String?> getLiveStreamURL(int id) async {
    var result = await api.getCameraLiveStream(id: id, redirect: false);

    return result.data?.uri;
  }
}

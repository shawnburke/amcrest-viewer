import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:amcrest_viewer_flutter/view_model/loading.viewmodel.dart';
import 'package:flutter/foundation.dart';
import 'package:openapi/api.dart';

import '../config.dart';

class HomeViewModel extends LoadingViewModel {
  HomeViewModel({
    required this.repo,
  });

  final CamViewerRepo repo;
  List<Camera>? _cameras;

  List<Camera> get cameras {
    return _cameras ?? List<Camera>.empty();
  }

  List<String> get snapshotUrls {
    return cameras
        .map((e) => baseURL + (e.latestSnapshot?.path ?? ''))
        .toList()
        .cast<String>();
  }

  Future<void> refresh() async {
    try {
      _cameras = await repo.getCameras(includeLastSnapshot: true);
      notifyListeners();
    } catch (exc) {
      debugPrint('Error in _fetchData : ${exc.toString()}');
    }
  }
}

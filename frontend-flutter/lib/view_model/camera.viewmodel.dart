import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:amcrest_viewer_flutter/view_model/loading.viewmodel.dart';
import 'package:openapi/api.dart';

import '../config.dart';

class CameraViewModel extends LoadingViewModel {
  final int cameraID;
  final CamViewerRepo repo;
  Camera? camera;

  CameraViewModel({
    required this.cameraID,
    required this.repo,
  });

  // factory CameraViewModel.withParameters(int cameraID, CamViewerRepo repo) {
  //   return CameraViewModel(cameraID: cameraID, repo: repo);
  // }

  get title {
    return camera?.name ?? '';
  }

  get url {
    String imagePath = "";
    if (camera?.latestSnapshot?.path != null) {
      imagePath = baseURL + (camera?.latestSnapshot?.path ?? '');
    }
    return imagePath;
  }

  void refresh() async {
    try {
      super.isLoading = true;
      camera = await repo.getCamera(cameraID);
    } finally {
      super.isLoading = false;
    }
  }
}

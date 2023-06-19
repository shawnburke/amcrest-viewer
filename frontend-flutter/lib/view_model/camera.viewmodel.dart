import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:amcrest_viewer_flutter/view_model/loading.viewmodel.dart';
import 'package:openapi/api.dart';

import '../config.dart';
import '../widgets/camera_widget.dart';

class CameraViewModel extends LoadingViewModel {
  final int cameraID;
  final CamViewerRepo repo;
  Camera? camera;
  List<CameraFile> files = List<CameraFile>.empty();

  List<CameraFile> get videos {
    return files.where((element) => element.type == 1).toList();
  }

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
      imagePath = CameraWidget.getImageURL(camera?.latestSnapshot?.path ?? '');
    }
    return imagePath;
  }

  void setRange(DateTime start, [DateTime? end]) async {
    end ??= start.add(const Duration(days: 1));

    try {
      super.isLoading = true;
      files = await repo.getFiles(cameraID, start, end);
      if (files.length > 500) {
        files = files.sublist(0, 500);
      }
    } finally {
      super.isLoading = false;
    }
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

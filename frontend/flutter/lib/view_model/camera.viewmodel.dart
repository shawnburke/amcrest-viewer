import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:amcrest_viewer_flutter/view_model/loading.viewmodel.dart';
import 'package:openapi/api.dart';

import '../config.dart';
import '../widgets/camera_widget.dart';

const maxFiles = -1;
const typeVideo = 1;
const typeImage = 0;

class CameraViewModel extends LoadingViewModel {
  final int cameraID;
  final CamViewerRepo repo;
  Camera? camera;
  List<CameraFile> files = List<CameraFile>.empty();
  List<CameraVideo>? _videoFiles;

  List<CameraVideo> get videos {
    return ensureVideos();
  }

  CameraViewModel({
    required this.cameraID,
    required this.repo,
  });

  List<CameraVideo> ensureVideos() {
    if (_videoFiles == null) {
      final vids = <CameraVideo>[];

      for (var i = 0; i < files.length; i++) {
        final file = files[i];
        if (file.type == typeVideo) {
          final thumbnail = files
              .sublist(i + 1)
              .firstWhere((element) => element.type == typeImage);
          vids.add(CameraVideo(file, thumbnail));
        }
      }
      _videoFiles = vids;
    }
    return _videoFiles!;
  }

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
      final ds = DateTime.now();
      files = await repo.getFiles(cameraID, start, end);
      if (files.length > maxFiles && maxFiles != -1) {
        files = files.sublist(0, maxFiles);
      }
      _videoFiles = null;
      print(
          'Loaded ${files.length} files in ${DateTime.now().difference(ds).inMilliseconds}ms.');
    } finally {
      super.isLoading = false;
    }
  }

  void refresh() async {
    try {
      super.isLoading = true;
      camera = await repo.getCamera(cameraID);

      if (files.isEmpty) {
        final now = DateTime.now();
        setRange(DateTime(now.year, now.month, now.day));
      }
    } finally {
      super.isLoading = false;
    }
  }
}

class CameraVideo {
  final CameraFile video;
  final CameraFile? thumbnail;

  CameraVideo(this.video, this.thumbnail);
}

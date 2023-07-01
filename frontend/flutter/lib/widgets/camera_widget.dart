import 'package:flutter/material.dart';
import 'package:openapi/api.dart';

import '../config.dart';
import 'camera_live_widget.dart';

class CameraWidget extends StatelessWidget {
  const CameraWidget({super.key, required this.camera});

  final Camera camera;

  Widget _buildCameraWidget(BuildContext context, {bool live = false}) {
    if (!live) {
      final url = getImageURL(camera.latestSnapshot?.path ?? '');
      return Image.network(url, width: 500);
    }

    return CameraVideoWidget(camera: camera);
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
        onTap: () => Navigator.pushNamed(context, 'camera/${camera.id}'),
        child: Center(
            child: Container(
                margin: const EdgeInsets.all(10.0),
                color: Colors.lightBlue[600],
                child: Column(children: [
                  Text(
                    camera.name,
                    textAlign: TextAlign.center,
                    textScaleFactor: 2,
                  ),
                  _buildCameraWidget(context, live: false),
                  // CameraVideoWidget(camera: camera),
                ]))));
  }

  static String getImageURL(String path) {
    if (path.startsWith("http")) {
      return path;
    }
    return Config.baseURL + path;
  }
}

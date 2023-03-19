import 'package:amcrest_viewer_flutter/widgets/camera_live_widget.dart';
import 'package:flutter/material.dart';
import 'package:openapi/api.dart';

import '../config.dart';

class CameraWidget extends StatelessWidget {
  const CameraWidget({super.key, required this.camera});

  final Camera camera;

  @override
  Widget build(BuildContext context) {
    final url = baseURL + (camera.latestSnapshot?.path ?? '');

    return GestureDetector(
        onTap: () =>
            Navigator.pushNamed(context, '/camera', arguments: camera.id),
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
                  Image.network(url, width: 500),
                  CameraVideoWidget(camera: camera),
                ]))));
  }
}

import 'dart:async';
import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:flutter/material.dart';
import 'package:openapi/api.dart';
import 'package:video_player/video_player.dart';

import '../config.dart';
import '../locator.dart';

class CameraVideoWidget extends StatefulWidget {
  final Camera camera;
  late final CamViewerRepo? repo;
  CameraVideoWidget({super.key, required this.camera, CamViewerRepo? repo}) {
    this.repo = repo ?? locator<CamViewerRepo>();
  }

  @override
  State<CameraVideoWidget> createState() => _VideoPlayerScreenState();
}

class _VideoPlayerScreenState extends State<CameraVideoWidget> {
  VideoPlayerController? _controller;
  Future<void>? _initFuture;
  Camera? camera;

  @override
  void initState() {
    super.initState();
    camera ??= widget.camera;

    _initFuture = widget.repo!.getLiveStreamURL(camera!.id).then((value) {
      // Create and store the VideoPlayerController. The VideoPlayerController
      // offers several different constructors to play videos from assets, files,
      // or the internet.
      _controller = VideoPlayerController.network(
        value!,
      );
      return _controller!.initialize();
    }).then((value) {
      setState(() {
        _controller!.setLooping(true);
      });
    });
  }

  @override
  void dispose() {
    // Ensure disposing of the VideoPlayerController to free up resources.
    _controller?.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder(
      future: _initFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.done) {
          // If the VideoPlayerController has finished initialization, use
          // the data it provides to limit the aspect ratio of the video.
          return AspectRatio(
            aspectRatio: _controller!.value.aspectRatio,
            // Use the VideoPlayer widget to display the video.
            child: VideoPlayer(_controller!),
          );
        } else {
          // If the VideoPlayerController is still initializing, show a
          // loading spinner.
          return const Center(
            child: CircularProgressIndicator(),
          );
        }
      },
    );
  }
}

import 'dart:async';
import 'package:flutter/material.dart';
import 'package:openapi/api.dart';
import 'package:video_player/video_player.dart';

import '../config.dart';

class CameraVideoWidget extends StatefulWidget {
  final Camera camera;
  const CameraVideoWidget({super.key, required this.camera});

  @override
  State<CameraVideoWidget> createState() => _VideoPlayerScreenState();
}

class _VideoPlayerScreenState extends State<CameraVideoWidget> {
  late VideoPlayerController _controller;
  late Future<void> _initializeVideoPlayerFuture;
  Camera? camera;

  @override
  void initState() {
    super.initState();
    camera ??= widget.camera;

    final url =
        '${Config.baseURL}/api/cameras/${camera!.id}/live?redirect=false';

    // Create and store the VideoPlayerController. The VideoPlayerController
    // offers several different constructors to play videos from assets, files,
    // or the internet.
    _controller = VideoPlayerController.network(
      url,
    );

    // Initialize the controller and store the Future for later use.
    _initializeVideoPlayerFuture = _controller.initialize();

    // Use the controller to loop the video.
    _controller.setLooping(true);
  }

  @override
  void dispose() {
    // Ensure disposing of the VideoPlayerController to free up resources.
    _controller.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder(
      future: _initializeVideoPlayerFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.done) {
          // If the VideoPlayerController has finished initialization, use
          // the data it provides to limit the aspect ratio of the video.
          return AspectRatio(
            aspectRatio: _controller.value.aspectRatio,
            // Use the VideoPlayer widget to display the video.
            child: VideoPlayer(_controller),
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

//import 'package:flick_video_player/flick_video_player.dart';
import 'package:chewie/chewie.dart';
import 'package:flutter/material.dart';
import 'package:video_player/video_player.dart';

import 'dart:developer' as developer;

class CameraImageWidget extends StatefulWidget {
  final String? imageURL;
  const CameraImageWidget(this.imageURL, {super.key});

  @override
  State<CameraImageWidget> createState() => CameraImageWidgetState();
}

class CameraImageWidgetState extends State<CameraImageWidget> {
  // FlickManager? _fvm;
  String? _imageURL;

  String? get imageURL => _imageURL;

  bool _setImageURL(String? newURL) {
    if (newURL != _imageURL) {
      _imageURL = newURL;
      return true;
    }
    return false;
  }

  @override
  void dispose() {
    //   _fvm?.dispose();

    _setImageURL(null);
    super.dispose();
  }

  static String _imageType(String? url) {
    if (url == null) {
      return 'none';
    }

    switch (url.split('.').last) {
      case 'jpg':
      case 'jpeg':
        return 'image';
      case 'mp4':
        return 'video';
      case 'm3u8':
        return 'stream';
      default:
        throw Exception('Unknown image type $url}');
    }
  }

  Widget _videoWidget(BuildContext context, {stream = false}) {
    VideoPlayerController controller =
        VideoPlayerController.networkUrl(Uri.parse(imageURL!));

    // if (_fvm == null) {
    //   _fvm = FlickManager(
    //     videoPlayerController: controller,
    //     autoInitialize: false,
    //     autoPlay: false,
    //   );
    // } else {
    //   _fvm!.handleChangeVideo(controller);
    // }
    // _fvm!.registerContext(context);

    // controller.initialize().then((_) {
    //   controller.play();
    //   controller.setVolume(0);
    //   // if (stream) {
    //   //   controller.setLooping(true);
    //   // }
    // }).onError((error, stackTrace) {
    //   developer.log('Error initializing video player:\n $error',
    //       error: error, stackTrace: stackTrace);
    // });

    final chewieController = ChewieController(
      videoPlayerController: controller,
      autoPlay: true,
      looping: stream,
    );

    chewieController.setVolume(0);

    return Chewie(
      controller: chewieController,
    );
  }

  Widget _buildWidget(BuildContext context) {
    try {
      switch (_imageType(imageURL)) {
        case 'none':
          return const Text('No image');
        case 'image':
          return Image.network(imageURL!);
        case 'video':
          return _videoWidget(context);
        case 'stream':
          return _videoWidget(context, stream: true);
        default:
          return const Text('Unknown image type');
      }
    } on Exception catch (e) {
      return Text('Error:\nURL=$imageURL\n$e');
    }
  }

  @override
  Widget build(BuildContext context) {
    _setImageURL(this.widget.imageURL);
    Widget? widget = _buildWidget(context);

    return Container(
      height: MediaQuery.of(context).size.height * 0.33,
      child: AspectRatio(aspectRatio: 16 / 9, child: widget),
    );
  }
}

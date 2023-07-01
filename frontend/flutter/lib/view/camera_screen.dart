import 'package:amcrest_viewer_flutter/locator.dart';
import 'package:amcrest_viewer_flutter/widgets/camera_widget.dart';
import 'package:flutter/material.dart';
import 'package:flutter_timeline_calendar/timeline/flutter_timeline_calendar.dart';
import 'package:intl/intl.dart';
import 'package:openapi/api.dart';
import 'package:provider/provider.dart';
import 'package:video_player/video_player.dart';

import '../repository/cam_viewer_repository.dart';
import '../view_model/camera.viewmodel.dart';
import '../widgets/timeline_view.dart';

class CameraScreen extends StatefulWidget {
  final int cameraID;
  const CameraScreen({super.key, required this.cameraID});

  @override
  State<CameraScreen> createState() => _CameraScreenState();
}

class _CameraScreenState extends State<CameraScreen> {
  VideoPlayerController? _controller;
  late final CameraViewModel vm;
  CalendarDateTime? _selectedDate;
  CameraVideo? _selectedVideo;
  CameraFile? _selectedFile;

  _CameraScreenState();

  @override
  void initState() {
    super.initState();
    vm = CameraViewModel(
        repo: locator<CamViewerRepo>(), cameraID: widget.cameraID);
    vm.refresh();
  }

  void _setActiveVideo(CameraVideo? vid) {
    if (_controller != null) {
      _controller!.pause();
      _controller?.dispose();
      _controller = null;
    }
    _selectedVideo = vid;
    _selectedFile = null;

    if (vid != null) {
      _controller = VideoPlayerController.network(
          CameraWidget.getImageURL(vid.video.path))
        ..initialize().then((_) {
          _controller!.play();
          _controller!.setVolume(0);
          // Ensure the first frame is shown after the video is initialized, even before the play button has been pressed.
          setState(() {});
        });
    }
  }

  @override
  void dispose() {
    super.dispose();
    if (_controller != null) {
      _controller!.pause();
      _controller?.dispose();
    }
  }

  Widget _buildVideoPlayer(CameraVideo file, {bool thumbnail = false}) {
    if (thumbnail && file.thumbnail != null) {
      return Image.network(CameraWidget.getImageURL(file.thumbnail!.path));
    }

    final controller = VideoPlayerController.network(
      CameraWidget.getImageURL(file.video.path),
    );

    controller.initialize();

    return VideoPlayer(controller);
  }

  Widget get _videoWidget {
    if (_selectedFile != null) {
      return Image.network(CameraWidget.getImageURL(_selectedFile!.path));
    }
    final initialized = _controller != null && _controller!.value.isInitialized;
    return Stack(
      fit: StackFit.expand,
      children: [
        initialized
            ? AspectRatio(
                aspectRatio: _controller!.value.aspectRatio,
                child: VideoPlayer(_controller!))
            : Container(),
        Positioned(
            bottom: 25,
            height: 50,
            width: MediaQuery.of(context).size.width,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                ElevatedButton(
                  onPressed: () {
                    setState(() {
                      if (_controller == null) {
                        return;
                      }
                      _controller!.value.isPlaying
                          ? _controller!.pause()
                          : _controller!.play();
                    });
                  },
                  child: Icon(
                    _controller != null && _controller!.value.isPlaying
                        ? Icons.pause
                        : Icons.play_arrow,
                  ),
                ),
                ElevatedButton(
                  onPressed: () {
                    setState(() {
                      if (_controller == null) {
                        return;
                      }
                      var vol = _controller!.value.volume;
                      if (vol == 0) {
                        vol = 1.0;
                      } else {
                        vol = 0;
                      }

                      _controller!.setVolume(vol);
                    });
                  },
                  child: Icon(
                    _controller != null && _controller!.value.volume > 0
                        ? Icons.volume_up
                        : Icons.volume_mute,
                  ),
                )
              ],
            )),
        Positioned(
            bottom: 0,
            height: 20,
            width: MediaQuery.of(context).size.width,
            child: initialized
                ? VideoProgressIndicator(
                    _controller!,
                    allowScrubbing: true,
                    colors: const VideoProgressColors(
                        backgroundColor: Colors.blueGrey,
                        bufferedColor: Colors.blueGrey,
                        playedColor: Colors.blueAccent),
                  )
                : Container()),
      ],
    );
  }

  Widget _buildListCell(CameraVideo file) {
    final formatted =
        DateFormat('E hh:mm a').format(file.video.timestamp.toLocal());

    final isSelected = _selectedVideo?.video.id == file.video.id;
    return Container(
        color:
            isSelected ? Colors.blueGrey.withOpacity(.5) : Colors.transparent,
        padding: const EdgeInsets.all(10.0),
        child: Column(
          children: [
            Text(
              '$formatted (${Duration(seconds: file.video.durationSeconds).inSeconds}s)',
              textScaleFactor: 1.0,
            ),
            SizedBox(
                width: 250, child: _buildVideoPlayer(file, thumbnail: true)),
          ],
        ));
  }

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider<CameraViewModel>(create: (context) {
      return vm;
    }, child: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: Consumer<CameraViewModel>(builder: (context, model, child) {
      return Scaffold(
          appBar: AppBar(
            // Here we take the value from the CameraScreen object that was created by
            // the App.build method, and use it to set our appbar title.
            title: Text('Camera Viewer (${vm.camera?.name ?? "missing"})'),
          ),
          body: Center(
            child: Container(
                margin: const EdgeInsets.all(5.0),
                color: Colors.lightBlue[600],
                child: Column(children: [
                  Expanded(
                    child: _videoWidget,
                  ),
                  TimelineCalendar(
                      dateTime: _selectedDate,
                      calendarType: CalendarType.GREGORIAN,
                      calendarLanguage: "en",
                      calendarOptions: CalendarOptions(
                        viewType: ViewType.DAILY,
                        toggleViewType: true,
                        headerMonthElevation: 10,
                        headerMonthShadowColor: Colors.black26,
                        headerMonthBackColor: Colors.transparent,
                      ),
                      dayOptions: DayOptions(
                          compactMode: true,
                          weekDaySelectedColor: const Color(0xff3AC3E2)),
                      headerOptions: HeaderOptions(
                          weekDayStringType: WeekDayStringTypes.SHORT,
                          monthStringType: MonthStringTypes.FULL,
                          backgroundColor: const Color(0xff3AC3E2),
                          headerTextColor: Colors.black),
                      onChangeDateTime: (datetime) {
                        _selectedDate = datetime;
                        vm.setRange(datetime.toDateTime());
                      }),
                  SizedBox(
                      height: 50,
                      child: TimelineView(vm.timelineItems, onTapped: (items) {
                        final vid = items
                            .where((element) => element.item is CameraVideo);

                        if (vid.isNotEmpty) {
                          _setActiveVideo(vid.first.item as CameraVideo);
                        } else if (items.isNotEmpty) {
                          setState(() {
                            _setActiveVideo(null);
                            _selectedFile = items.first.item as CameraFile?;
                          });
                        }
                      })),
                  Expanded(
                      child: ListView.builder(
                    shrinkWrap: true,
                    padding: const EdgeInsets.all(8),
                    itemCount: vm.videos.length,
                    scrollDirection: Axis.horizontal,
                    itemBuilder: (BuildContext context, int index) {
                      return GestureDetector(
                        onTap: () => _setActiveVideo(vm.videos[index]),
                        child: _buildListCell(vm.videos[index]),
                      );
                    },
                  )),
                ])),
          ));
    })));
  }
}

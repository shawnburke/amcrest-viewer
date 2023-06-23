import 'package:amcrest_viewer_flutter/locator.dart';
import 'package:amcrest_viewer_flutter/widgets/camera_widget.dart';
import 'package:flutter/material.dart';
import 'package:flutter_timeline_calendar/timeline/flutter_timeline_calendar.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:video_player/video_player.dart';

import '../repository/cam_viewer_repository.dart';
import '../view_model/camera.viewmodel.dart';

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

  _CameraScreenState();

  @override
  void initState() {
    super.initState();
    vm = CameraViewModel(
        repo: locator<CamViewerRepo>(), cameraID: widget.cameraID);
  }

  void _setActiveVideo(CameraVideo vid) {
    _controller =
        VideoPlayerController.network(CameraWidget.getImageURL(vid.video.path))
          ..initialize().then((_) {
            _controller!.play();
            _controller!.setVolume(0);
            // Ensure the first frame is shown after the video is initialized, even before the play button has been pressed.
            setState(() {});
          });
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
    final initialized = _controller != null && _controller!.value.isInitialized;
    return Stack(
      children: [
        initialized
            ? AspectRatio(
                aspectRatio: _controller!.value.aspectRatio,
                child: VideoPlayer(_controller!))
            : Container(),
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

  Widget _buildListRow(CameraVideo file) {
    final formatted = DateFormat('E hh:mm a').format(file.video.timestamp);
    return Container(
        padding: const EdgeInsets.all(10.0),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: [
            Text('$formatted\n(${file.video.durationSeconds}s)'),
            _buildVideoPlayer(file, thumbnail: true),
          ],
        ));
  }

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider<CameraViewModel>(create: (context) {
      vm.refresh();
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
                    Text(
                      vm.title,
                      textAlign: TextAlign.center,
                      textScaleFactor: 2,
                    ),
                    SizedBox(height: 300, child: _videoWidget),
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
                    Expanded(
                      child: ListView.builder(
                        shrinkWrap: true,
                        padding: const EdgeInsets.all(8),
                        itemCount: vm.videos.length,
                        scrollDirection: Axis.vertical,
                        itemBuilder: (BuildContext context, int index) {
                          return SizedBox(
                            height: 200,
                            width: 310,
                            child: GestureDetector(
                              onTap: () => _setActiveVideo(vm.videos[index]),
                              child: _buildListRow(vm.videos[index]),
                            ),
                          );
                        },
                      ),
                    ),
                  ]))));
    })));
  }
}

// class _CameraScreenState extends State<CameraScreen> {
//   late CameraViewModel _camViewModel;
//   final String _cameraID;

//   _CameraScreenState(String cameraID) {
//     _cameraID = cameraID;
//   }

//   @override
//   void initState() {
//     _camViewModel = Provider.of<CameraViewModel>(context, listen: false);
//     _camViewModel.cameraID = _cameraID;
//     super.initState();
//     _camViewModel.refresh();
//   }

//   @override
//   Widget build(BuildContext context) {
//     return Scaffold(
//       appBar: AppBar(
//         // Here we take the value from the CameraScreen object that was created by
//         // the App.build method, and use it to set our appbar title.
//         title: Text(widget.title),
//       ),
//       body: Center(
//         // Center is a layout widget. It takes a single child and positions it
//         // in the middle of the parent.
//         child: Consumer<CameraViewModel>(
//           builder: (context, model, child) {
//             return Center(
//                 child: Container(
//                     margin: const EdgeInsets.all(10.0),
//                     color: Colors.lightBlue[600],
//                     child: Column(children: [
//                       Text(
//                         camera.name,
//                         textAlign: TextAlign.center,
//                         textScaleFactor: 2,
//                       ),
//                       Image.network(url, width: 500),
//                     ])));
//           },
//         ),
//       ),
//     );
//   }
// }

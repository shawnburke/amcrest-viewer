import 'package:amcrest_viewer_flutter/locator.dart';
import 'package:amcrest_viewer_flutter/widgets/camera_widget.dart';
import 'package:flutter/material.dart';
import 'package:flutter_timeline_calendar/timeline/flutter_timeline_calendar.dart';
import 'package:openapi/api.dart';
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
  late final Widget _loadingWidget;

  _CameraScreenState();

  @override
  void initState() {
    super.initState();
    vm = CameraViewModel(
        repo: locator<CamViewerRepo>(), cameraID: widget.cameraID);
    vm.refresh();
    _loadingWidget = const Center(child: CircularProgressIndicator());
  }

  void _setActiveVideo(CameraFile vid) {
    _controller =
        VideoPlayerController.network(CameraWidget.getImageURL(vid.path))
          ..initialize().then((_) {
            _controller!.play();
            // Ensure the first frame is shown after the video is initialized, even before the play button has been pressed.
            setState(() {});
          });
  }

  Widget _buildVideoPlayer(CameraFile file) {
    final controller = VideoPlayerController.network(
      CameraWidget.getImageURL(file.path),
    );

    final widget = VideoPlayer(controller);

    return widget;
  }

  Widget get _videoWidget {
    return _controller == null ? _loadingWidget : VideoPlayer(_controller!);
  }

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider<CameraViewModel>(
        create: (context) => vm,
        child: Center(
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
                      margin: const EdgeInsets.all(10.0),
                      color: Colors.lightBlue[600],
                      child: Column(children: [
                        Text(
                          vm.title,
                          textAlign: TextAlign.center,
                          textScaleFactor: 2,
                        ),
                        Expanded(child: _videoWidget),
                        TimelineCalendar(
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
                                    onTap: () =>
                                        _setActiveVideo(vm.videos[index]),
                                    child: _buildVideoPlayer(vm.videos[index])),
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

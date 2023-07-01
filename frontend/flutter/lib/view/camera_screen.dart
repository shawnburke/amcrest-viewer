import 'package:amcrest_viewer_flutter/locator.dart';
import 'package:amcrest_viewer_flutter/widgets/camera_widget.dart';
import 'package:flick_video_player/flick_video_player.dart';
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
    if (_flickManager != null) {
      _flickManager!.dispose();
      _flickManager = null;
    }

    setState(() {
      _selectedVideo = vid;
      _selectedFile = null;
    });
  }

  @override
  void dispose() {
    super.dispose();
    _setActiveVideo(null);
  }

  FlickManager? _flickManager;

  FlickManager? getFlickManager() {
    if (_selectedVideo == null) {
      return null;
    }

    if (_flickManager != null) {
      return _flickManager;
    }

    VideoPlayerController? controller;
    controller = VideoPlayerController.network(
        CameraWidget.getImageURL(_selectedVideo!.video.path));

    _flickManager = FlickManager(
      videoPlayerController: controller,
    );

    return _flickManager;
  }

  Widget get _videoWidget {
    if (_selectedFile != null) {
      return Expanded(
          child: Image.network(CameraWidget.getImageURL(_selectedFile!.path)));
    }

    final mgr = getFlickManager();

    if (mgr == null) {
      return Container(height: MediaQuery.of(context).size.height * 0.2);
    }

    return Expanded(
        child: Stack(
      fit: StackFit.loose,
      children: [
        FlickVideoPlayer(flickManager: mgr),
      ],
    ));
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
                width: 250,
                child: Image.network(
                    CameraWidget.getImageURL(file.thumbnail!.path))),
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
                  _videoWidget,
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
                  Container(
                      height: 50,
                      child: TimelineView(vm.timelineItems, onTapped: (items) {
                        final vid = items
                            .where((element) => element.item is CameraVideo);

                        if (vid.isNotEmpty) {
                          _setActiveVideo(vid.first.item as CameraVideo);
                          return;
                        }

                        if (items.isNotEmpty) {
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

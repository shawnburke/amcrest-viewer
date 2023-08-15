import 'package:amcrest_viewer_flutter/locator.dart';
import 'package:amcrest_viewer_flutter/widgets/camera_image_widget.dart';
import 'package:amcrest_viewer_flutter/widgets/camera_widget.dart';
import 'package:flutter/material.dart';
import 'package:flutter_timeline_calendar/timeline/flutter_timeline_calendar.dart';
import 'package:intl/intl.dart';
import 'package:openapi/openapi.dart';
import 'package:provider/provider.dart';

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
  String? _media;

  _CameraScreenState();

  @override
  void initState() {
    super.initState();
    vm = CameraViewModel(
        repo: locator<CamViewerRepo>(), cameraID: widget.cameraID);
    vm.refresh();
  }

  void _setActiveVideo(CameraVideo? vid) {
    _setMedia(video: vid);
  }

  @override
  void dispose() {
    super.dispose();
    _setActiveVideo(null);
  }

  void _setMedia({
    CameraVideo? video,
    CameraFile? file,
    String? uri,
  }) {
    String? mediaURL;

    if (uri != null) {
      mediaURL = CameraWidget.getImageURL(uri);
    } else if (video != null) {
      mediaURL = CameraWidget.getImageURL(video.video.path);
    } else if (file != null) {
      mediaURL = CameraWidget.getImageURL(file.path);
    }

    setState(() {
      _selectedFile = file;
      _selectedVideo = video;
      _media = mediaURL;
    });
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
                  CameraImageWidget(_media),
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
                          _setMedia(video: vid.first.item as CameraVideo);
                          return;
                        }

                        if (items.isNotEmpty) {
                          setState(() {
                            _setMedia(file: items.first.item as CameraFile?);
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
                  ElevatedButton(
                      onPressed: () async {
                        final liveURL = await vm.liveURL;
                        if (liveURL == null) {
                          return;
                        }
                        _setMedia(uri: liveURL);
                      },
                      child: const Text("Live")),
                ])),
          ));
    })));
  }
}

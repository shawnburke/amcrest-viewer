import 'package:amcrest_viewer_flutter/view/camera_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter/widgets.dart';
import 'package:openapi/api.dart';

import '../view_model/camera.viewmodel.dart';

class TimelineView extends StatefulWidget {
  final List<TimelineItem> items;
  final Function(Iterable<TimelineItem> items)? onTapped;

  TimelineView(this.items, {this.onTapped, super.key});

  @override
  _TimelineViewState createState() => _TimelineViewState();
}

class TimelineViewModel extends ChangeNotifier {
  TimelineViewModel();

  void dispose() {
    super.dispose();
  }
}

class TimelineItem<T> {
  final DateTime time;
  final T item;

  TimelineItem({
    required this.time,
    required this.item,
  });
}

const durationMinute = Duration(minutes: 1);

class TimelineItemCollection {
  final List<TimelineItem<dynamic>> items;
  static const _bucketSize = Duration(minutes: 5);
  Map<DateTime, List<TimelineItem<dynamic>>> _buckets = {};

  DateTime get start => items.first.time.truncate(durationMinute);
  DateTime get end =>
      items.last.time.truncate(durationMinute).add(durationMinute);
  int get minutes => end.difference(start).inMinutes;

  TimelineItemCollection(this.items);

  void _ensureBuckets() {
    if (_buckets.isEmpty) {
      _buckets = {};
      for (var item in items) {
        final bucket = item.time.truncate(_bucketSize);
        if (!_buckets.containsKey(bucket)) {
          _buckets[bucket] = [];
        }
        _buckets[bucket]!.add(item);
      }
    }
  }

  List<TimelineItem<dynamic>> getItems(DateTime time) {
    _ensureBuckets();
    final bucket = time.truncate(_bucketSize);

    final start = time.truncate(durationMinute);
    final end = start.add(durationMinute);

    return _buckets[bucket]!
        .where((element) =>
            element.time.isAfter(start) && element.time.isBefore(end))
        .toList();
  }
}

extension DateTimeExtensions on DateTime {
  DateTime truncate(Duration d) {
    final unix = millisecondsSinceEpoch;

    final truncated = unix - (unix % d.inMilliseconds);

    return DateTime.fromMillisecondsSinceEpoch(truncated);
  }
}

class _TimelineViewState extends State<TimelineView> {
  final colors = [Colors.red, Colors.white, Colors.blue];
  TimelineItemCollection? _c;
  Function(Iterable<TimelineItem>)? _onTapped;

  _TimelineViewState();

  @override
  void didUpdateWidget(covariant TimelineView oldWidget) {
    _c = null;
    super.didUpdateWidget(oldWidget);
  }

  TimelineItemCollection get _collection {
    if (_c == null) {
      final sorted = widget.items..sort((a, b) => a.time.compareTo(b.time));
      _c = TimelineItemCollection(sorted);
    }
    return _c!;
  }

  void _fireTapped(Iterable<TimelineItem> items) {
    if (widget.onTapped != null) {
      widget.onTapped!(items);
    }
  }

  @override
  Widget build(BuildContext context) {
    final totalMinutes = _collection.minutes;

    return Positioned.fill(
        child: ListView.builder(
            itemCount: totalMinutes,
            scrollDirection: Axis.horizontal,
            itemBuilder: (context, index) {
              final time = _collection.start
                  .add(Duration(minutes: index))
                  .truncate(durationMinute);

              Widget? timeMarker = null;

              final isHour = time.minute == 0;
              if (isHour) {
                timeMarker = Container(
                    width: 10,
                    color: Colors.black,
                    child: Text(
                      time.hour.toString(),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 10,
                      ),
                    ));
              }
              // final isHalfHour = time.minute == 30;
              // final isQuarterHour = !isHour && !isHalfHour && time.minute % 15 == 0;

              final items = _collection.getItems(time);
              final images = items
                  .where((element) => element.item is CameraFile)
                  .map((e) => e.item as CameraFile)
                  .toList();

              final videos = items
                  .where((element) => element.item is CameraVideo)
                  .map((e) => e.item as CameraVideo)
                  .toList();

              var width = 2.0;
              var color = Colors.grey;
              if (videos.isNotEmpty) {
                color = Colors.yellow;
                width = 5;
              } else if (images.isNotEmpty) {
                color = Colors.lightGreen;
              }
              final widget = GestureDetector(
                  onTap: () => _fireTapped(items),
                  child: Container(
                    color: color,
                    width: width,
                  ));

              if (timeMarker != null) {
                return Column(
                  children: [timeMarker, widget],
                );
              }
              return widget;
            }));
  }
}

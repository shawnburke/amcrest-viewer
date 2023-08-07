import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:openapi/openapi.dart';
import 'dart:developer' as developer;

import '../view_model/camera.viewmodel.dart';

class TimelineView extends StatefulWidget {
  final List<TimelineItem> items;
  final Function(Iterable<TimelineItem> items)? onTapped;

  const TimelineView(this.items, {this.onTapped, super.key});

  @override
  TimelineViewState createState() => TimelineViewState();
}

class TimelineViewModel extends ChangeNotifier {
  TimelineViewModel();
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

  DateTime get start => items.isEmpty
      ? DateTime.now()
      : items.first.time.truncate(durationMinute);
  DateTime get end => items.isEmpty
      ? DateTime.now()
      : items.last.time.truncate(durationMinute).add(durationMinute);
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
    final s = DateTime.now();
    try {
      _ensureBuckets();
      final bucket = time.truncate(_bucketSize);

      final start = time.truncate(durationMinute);
      final end = start.add(durationMinute);

      final b = _buckets[bucket];
      if (b == null) {
        developer.log('no bucket for $bucket');
        return [];
      }

      return b
          .where((element) =>
              element.time.isAfter(start) && element.time.isBefore(end))
          .toList();
    } finally {
      final ms = DateTime.now().difference(s).inMilliseconds;
      if (ms > 5) {
        developer.log('getItems took $ms ms');
      }
    }
  }

  String info(DateTime time) {
    _ensureBuckets();
    final bucket = time.truncate(_bucketSize);

    final start = time.truncate(durationMinute);
    final end = start.add(durationMinute);

    var str = 'bucket: $bucket, start: $start, end: $end';
    final b = _buckets[bucket];
    if (b == null) {
      str += ' -- no bucket';
    } else {
      str += 'bucket: ${b.length}';
    }
    return str;
  }
}

extension DateTimeExtensions on DateTime {
  DateTime truncate(Duration d) {
    final unix = millisecondsSinceEpoch;

    final truncated = unix - (unix % d.inMilliseconds);

    return DateTime.fromMillisecondsSinceEpoch(truncated);
  }
}

class TimelineViewState extends State<TimelineView> {
  final colors = [Colors.red, Colors.white, Colors.blue];
  TimelineItemCollection? _c;
  Function(Iterable<TimelineItem>)? _onTapped;

  TimelineViewState();

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
    var printed = false;

    return Container(
        color: Colors.black,
        child: ListView.builder(
            itemCount: totalMinutes,
            scrollDirection: Axis.horizontal,
            itemBuilder: (context, index) {
              final time = _collection.start
                  .add(Duration(minutes: index))
                  .truncate(durationMinute);

              Widget? timeMarker;

              final isHour = time.minute == 0;
              if (isHour) {
                timeMarker = Container(
                    width: 15,
                    color: Colors.black,
                    child: Text(
                      DateFormat("ha").format(time).replaceAll("M", ""),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 10,
                      ),
                    ));
              }

              final items = _collection.getItems(time);
              //print('time: $time, items: ${items.length}');

              final ti = _TimelineItem(items, time, key: GlobalKey());
              Widget widget = ti;

              if (items.isNotEmpty) {
                widget =
                    GestureDetector(onTap: () => _fireTapped(items), child: ti);
              }

              if (timeMarker != null) {
                return Column(
                  children: [timeMarker, widget],
                );
              }
              return widget;
            }));
  }
}

class _TimelineItem extends StatelessWidget {
  final DateTime time;
  final List<TimelineItem> items;

  const _TimelineItem(this.items, this.time, {super.key});

  @override
  Widget build(BuildContext context) {
    final images = items
        .where((element) => element.item is CameraFile)
        .map((e) => e.item as CameraFile)
        .toList();

    final videos = items
        .where((element) => element.item is CameraVideo)
        .map((e) => e.item as CameraVideo)
        .toList();

    var width = 2.0;
    var color = Colors.transparent;
    if (videos.isNotEmpty) {
      color = Colors.yellow;
      width = 5;
    } else if (images.isEmpty) {
      color = Colors.grey;

      // if (!printed) {
      //   printed = true;
      //   print('no items for $time, info: ${_collection.info(time)}');
      // }
    }
    return Container(
      color: color,
      width: width,
    );
  }
}

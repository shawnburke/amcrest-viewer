import 'dart:io';
import 'package:ansicolor/ansicolor.dart';
import 'package:logging/logging.dart';

class LogManager {
  final nowLogger = Logger('now');
  final bool bufferLogs;
  var _logs = <LogRecord>[];
  var _components = Set<String>();

  LogManager(Level level, {this.bufferLogs = false}) {
    Logger.root.level = level;
    Logger.root.onRecord.listen(_onLogRecord);
  }

  Logger get log => Logger.root;

  Logger get now => nowLogger;

  Iterable<String> get components => _components;

  void _onLogRecord(LogRecord record) {
    final now = !bufferLogs ||
        record.loggerName == nowLogger.name ||
        record.level >= Level.SEVERE;

    if (now) {
      print(_format(record));
      return;
    }
    _components.add(record.loggerName);
    _logs.add(record);
  }

  Iterable<String> get({String? componentName, DateTime? after}) {
    final buffer = <String>[];
    final newLogs = <LogRecord>[];
    for (var i = 0; i < _logs.length; i++) {
      final record = _logs[i];

      final isAfter = after == null || record.time.isAfter(after);
      final isTarget =
          componentName == null || record.loggerName == componentName;

      bool output = isTarget && isAfter;

      if (output) {
        buffer.add(_format(record));
      }

      if (!output) {
        newLogs.add(record);
      }
    }
    _logs = newLogs;
    return buffer;
  }

  Future<void> write(
      {String? componentName,
      bool stderr = false,
      DateTime? after,
      bool Function()? until}) async {
    while (true) {
      for (var line in get(componentName: componentName, after: after)) {
        stdout.writeln(line);
      }
      await Future.delayed(Duration(milliseconds: 100));

      if (until == null || !until()) {
        break;
      }
    }
  }

  String _format(LogRecord record) {
    final pen = getPenForLevel(record.level);
    final grey = AnsiPen()..gray();
    final component = record.loggerName.isEmpty ? '' : '|${record.loggerName}';

    return pen('[${record.level.name}]') +
        grey('$component') +
        ': ${record.message}';
  }

  AnsiPen getPenForLevel(Level level) {
    if (level == Level.INFO) {
      return AnsiPen()..blue();
    } else if (level == Level.WARNING) {
      return AnsiPen()..yellow();
    } else if (level == Level.SEVERE) {
      return AnsiPen()..red();
    } else if (level.value <= Level.FINE.value) {
      return AnsiPen()..green();
    } else {
      return AnsiPen()..white();
    }
  }
}

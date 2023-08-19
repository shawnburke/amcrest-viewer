import 'dart:async';
import 'dart:io';
import 'package:interact/interact.dart';
import 'package:mockoon_proxy/server/mock_data_server_client.dart';

import 'log_manager.dart';
import 'mock_data_server.dart';

class Interactive {
  final LogManager logger;
  final MockoonServerClient client;
  final void Function()? quit;
  String? currentScenario;
  var escCount = 0;

  static const ESC = 27;
  static const M = 109;
  static const L = 108;
  static const Q = 113;

  Interactive(this.logger, this.client, {this.quit, this.currentScenario});

  Future<void> show() async {
    stdin
      ..lineMode = false
      ..echoMode = false;

    stdin.listen((List<int> bytes) async {
      if (bytes[0] == ESC) {
        escCount++;
        return;
      }
      unawaited(dispatchCommand(bytes[0]));
    });

    await dispatchCommand(0);
  }

  Future<void> dispatchCommand(int command) async {
    switch (command) {
      case 0:
        break;
      case M:
        escCount++;
        await rootMenu();
        break;
      case Q:
        if (quit != null) {
          quit!();
          return;
        }
        exit(0);

      case L:
        final ec = escCount;
        await logger.write(
          until: () => ec == escCount,
        );
        break;

      default:
        print("unknown command: $command");
    }
    if (currentScenario != null) {
      print('Current scenario: $currentScenario\n');
    }
    print('Enter a command: M for menu, L for logs, Q to quit.\n');
  }

  Future<void> rootMenu() async {
    final baseOptions = [
      'choose_scenario|Choose scenario',
      'logs|View logs',
    ];

    final scenarioOptions = ['record|Record to', 'replay|Replay from'];

    var options = baseOptions;

    if (currentScenario != null) {
      options = scenarioOptions.map((e) => '$e $currentScenario').toList() +
          baseOptions;
    }

    final root = _SelectionItem('Choose option', options);

    final selection = await root.show();

    switch (selection) {
      case 'logs':
        await chooseLogger(logger);
        break;
      case 'choose_scenario':
        currentScenario =
            await chooseScenario(logger, client) ?? currentScenario;
        break;
      case 'record':
        await recordScenario(currentScenario!, logger, client);
        break;
      case 'replay':
        await replayScenario(currentScenario!, logger, client);
        break;
      default:
        logger.now.shout('Unknown: $selection');
        break;
    }
  }

  Future<void> recordScenario(
      String scenario, LogManager logger, MockoonServerClient client) async {
    final result = await client.setMode(mode: Mode.record, scenario: scenario);

    if (result['mode'] != Mode.record.name) {
      logger.now.severe('Failed to set mode to record');
      return;
    }
    print('\nRecording scenario $scenario, press ESC to stop.\n');
    final c = escCount;

    await logger.write(
      after: DateTime.now(),
      until: () => escCount == c,
    );

    print('\nClosing scenario and writing mockfile...');
    final close = await client.closeScenario(scenario: scenario);
    print(
        '\nScenario written to ${close['path']}.\nYou can now replay this scenario.\n');
  }

  Future<void> replayScenario(
      String scenario, LogManager logger, MockoonServerClient client,
      {bool fallthru = true}) async {
    final mode = fallthru ? Mode.replay_fallthrough : Mode.replay;

    final result = await client.setMode(mode: mode, scenario: scenario);

    if (result['mode'] != mode.name) {
      logger.now.severe('Failed to set mode to record');
      return;
    }
    print('\Replaying scenario $scenario, press ESC to stop.\n');
    final c = escCount;

    await logger.write(
      after: DateTime.now(),
      until: () => escCount == c,
    );
    print("\nReplaying scenario $scenario restopped.");
  }

  Future<String?> chooseScenario(
      LogManager logger, MockoonServerClient client) async {
    final scenarios = await client.getScenarios();
    var chooseScenario = _SelectionItem(
        'Choose scenario', ['new|Create new scenario...'] + scenarios);

    var chosen = await chooseScenario.show();
    if (chosen == 'new') {
      while (true) {
        var name = Input(prompt: 'Scenario name').interact();
        if (scenarios.contains(name)) {
          print('Scenario $name already exists');
          continue;
        }
        if (!name.startsWith('private')) {
          name = 'private-$name';
        }
        print(
            'New scenarios are prefiexed with "private" to prevent them from being accidentally committed to source control.\n' +
                'Rename to remove the prefix if you wish to commit.');
        chosen = name;
        break;
      }
    }
    return chosen;
  }

  Future<void> chooseLogger(LogManager logger) async {
    final log =
        _SelectionItem("Choose Logger", ['All'] + logger.components.toList());

    final selection = await log.show();

    final cn = selection == 'All' ? null : selection;

    print('Following logs for ${selection}, ESC to quit.');
    final c = escCount;
    await logger.write(
      componentName: cn,
      until: () => escCount == c,
    );
  }
}

class _SelectionItem {
  static const String separator = '|';
  final String prompt;
  final List<String> options;
  _SelectionItem(this.prompt, this.options);

  Future<String?> show() async {
    final menuOptions = options.map((e) => getLabel(e));
    final selection = Select(
      prompt: prompt,
      options: menuOptions.toList(),
    ).interact();

    for (final o in options) {
      if (o == selection || o.startsWith(o + separator)) {
        return o;
      }
    }

    return getKey(options[selection]);
  }

  String getLabel(String item) {
    final parts = item.split(separator);
    if (parts.length == 1) {
      return parts[0];
    }
    return parts[1];
  }

  String getKey(String item) {
    final parts = item.split(separator);
    if (parts.length == 1) {
      return parts[0];
    }
    return parts[0];
  }
}

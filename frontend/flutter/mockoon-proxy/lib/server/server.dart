import 'dart:io';
import 'package:args/args.dart';
import 'package:logging/logging.dart';
import 'package:mockoon_proxy/server/mock_data_server_client.dart';

import 'interactive.dart';
import 'log_manager.dart';
import 'mock_data_server.dart';

const argPort = 'port';
const argVerbose = 'verbose';
const argInteractive = 'interactive';
const argScenario = 'scenario';
const argMode = 'mode';
const argMockoonPort = 'mockoon-port';

Logger logger = Logger('main');

void main(List<String> arguments) async {
  final parser = ArgParser()
    ..addFlag(argVerbose,
        abbr: 'v',
        defaultsTo: false,
        negatable: false,
        help: 'Display verbose logs')
    ..addFlag(argInteractive,
        abbr: 'i',
        defaultsTo: false,
        negatable: false,
        help: 'Display interactive menu')
    ..addOption(argPort,
        abbr: 'p', defaultsTo: Platform.environment['PORT'] ?? '9099')
    ..addOption(argMode,
        abbr: 'm', defaultsTo: Mode.disabled.name, help: 'Mode to start in')
    ..addOption(argMockoonPort,
        defaultsTo: null,
        help:
            'Specify port from existing Mockoon server -- will not start instances.')
    ..addOption(argScenario,
        abbr: 's', defaultsTo: null, help: 'Scenario to load');

  try {
    parser.parse(arguments);
  } on FormatException {
    usage(parser);
  }

  ArgResults argResults = parser.parse(arguments);
  final host = Platform.environment['HOST'] ?? '0.0.0.0';
  final port = argResults[argPort];

  var level = Level.INFO;
  Logger.root.level = Level.INFO;
  if (argResults[argVerbose] == true) {
    level = Level.ALL;
  }
  final logger = LogManager(level, bufferLogs: argResults[argInteractive]);

  // look for an env variable called SCENARIO_DIR
  // if it exists, use that as the directory to store the scenarios
  // otherwise, use the default
  final scenarioDir = Platform.environment['SCENARIO_DIR'] ?? 'scenarios';

  final mode = Mode.values.firstWhere(
      (element) => element.name == argResults[argMode],
      orElse: () => Mode.disabled);

  final scenario = argResults[argScenario] as String?;

  final server = MockDataServer(int.parse(port), dir: scenarioDir, host: host);

  final mockoonPort = int.parse(argResults[argMockoonPort] ?? '0');

  // we allow passing in a static port so we can use a running mockoon.
  if (mockoonPort != 0) {
    if (scenario == null) {
      print('Cannot specify mockoon-port without scenario');
      usage(parser);
    }
    server.setScenarioPort(scenario!, mockoonPort);
  }

  void quit({int code = 0}) async {
    print('Cleaning up...');
    await server.dispose();
    exit(code);
  }

  final ctrlC = ProcessSignal.sigint;
  ctrlC.watch().listen((signal) async {
    quit();
  });

  await server.start();

  logger.write(componentName: 'MockDataServer');
  final client = MockoonServerClient(server.port, host: host);

  await client.setMode(mode: mode, scenario: scenario);

  if (argResults[argInteractive] == true) {
    final interactive =
        Interactive(logger, client, currentScenario: scenario, quit: quit);
    await interactive.show();
  }
}

void usage(ArgParser parser) {
  print("Mockoon Server Mocking Tool");
  print("Invalid command line arguments.");
  print(parser.usage);
  exit(1);
}

import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:amcrest_viewer_flutter/view/camera_screen.dart';
import 'package:amcrest_viewer_flutter/view/home_screen.dart';
import 'package:amcrest_viewer_flutter/view_model/home.viewmodel.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter_deep_linking/flutter_deep_linking.dart' as deep_linking;

import 'locator.dart';

void main() {
  setupLocator();
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(
          create: (_) => HomeViewModel(repo: locator<CamViewerRepo>()),
        ),
      ],
      child: const AmcrestViewer(),
    ),
  );
}

class AmcrestViewer extends StatelessWidget {
  const AmcrestViewer({super.key});

  int parseArguments(BuildContext context) {
    final args = ModalRoute.of(context)!.settings.arguments;
    if (args is Map) {
      return args['id'] as int;
    } else {
      return 0;
    }
  }

  RouteFactory buildRouter() {
    final router = deep_linking.Router(
      routes: [
        deep_linking.Route(
          matcher: deep_linking.Matcher.path('home'),
          builder: (result) => MaterialPageRoute(
            builder: (ctx) => const HomeScreen(),
          ),
        ),
        deep_linking.Route(
          matcher: deep_linking.Matcher.path('camera/{id}'),
          builder: (result) => MaterialPageRoute(
            builder: (ctx) => CameraScreen(
              cameraID: int.parse(result.parameters['id'] as String),
            ),
          ),
        ),
      ],
    );

    return router.onGenerateRoute;
  }

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Camera Viewer',
      initialRoute: 'home',
      onGenerateRoute: buildRouter(),
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
    );
  }
}

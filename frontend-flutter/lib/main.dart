import 'package:amcrest_viewer_flutter/repository/cam_viewer_repository.dart';
import 'package:amcrest_viewer_flutter/view/camera_screen.dart';
import 'package:amcrest_viewer_flutter/view/home_screen.dart';
import 'package:amcrest_viewer_flutter/view_model/home.viewmodel.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

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

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Camera Viewer',
      routes: {
        '/camera': (context) => CameraScreen(
            cameraID: ModalRoute.of(context)!.settings.arguments as int),
      },
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: const HomeScreen(),
    );
  }
}

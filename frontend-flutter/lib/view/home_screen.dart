import 'dart:async';

import 'package:amcrest_viewer_flutter/view_model/home.viewmodel.dart';
import 'package:flutter/material.dart';
import 'package:openapi/api.dart';
import 'package:provider/provider.dart';

import '../config.dart';
import '../widgets/camera_widget.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  final String title = 'Camera Viewer (Home)';

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _refreshSeconds = 15;

  late HomeViewModel _homeViewModel;

  _HomeScreenState({int refresh = 15}) {
    _refreshSeconds = refresh;
  }

  @override
  void initState() {
    _homeViewModel = Provider.of<HomeViewModel>(context, listen: false);

    super.initState();
    scheduleTimeout();
    _homeViewModel.refresh();
  }

  Timer scheduleTimeout() =>
      Timer(Duration(seconds: _refreshSeconds), handleTimeout);

  void handleTimeout() {
    _homeViewModel.refresh();
    scheduleTimeout();
  }

  @override
  Widget build(BuildContext context) {
    // This method is rerun every time setState is called, for instance as done
    // by the _updateImages method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    return Scaffold(
      appBar: AppBar(
        // Here we take the value from the HomeScreen object that was created by
        // the App.build method, and use it to set our appbar title.
        title: Text(widget.title),
      ),
      body: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: Consumer<HomeViewModel>(
          builder: (context, model, child) {
            return Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: _homeViewModel.cameras
                    .map((c) => CameraWidget(camera: c))
                    .toList()
                    .cast<Widget>());
          },
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _homeViewModel.refresh(),
        tooltip: 'Update',
        child: const Icon(Icons.add),
      ), // This trailing comma makes auto-formatting nicer for build methods.
    );
  }
}

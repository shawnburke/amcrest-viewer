import 'dart:async';

import 'package:amcrest_viewer_flutter/view_model/home.viewmodel.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

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
        child:
            // Column is also a layout widget. It takes a list of children and
            // arranges them vertically. By default, it sizes itself to fit its
            // children horizontally, and tries to be as tall as its parent.
            //
            // Invoke "debug painting" (press "p" in the console, choose the
            // "Toggle Debug Paint" action from the Flutter Inspector in Android
            // Studio, or the "Toggle Debug Paint" command in Visual Studio Code)
            // to see the wireframe for each widget.
            //
            // Column has various properties to control how it sizes itself and
            // how it positions its children. Here we use mainAxisAlignment to
            // center the children vertically; the main axis here is the vertical
            // axis because Columns are vertical (the cross axis would be
            // horizontal).
            Consumer<HomeViewModel>(
          builder: (context, model, child) {
            return Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: _homeViewModel.snapshotUrls
                  .map((url) => Image.network(url, width: 500))
                  .toList(),
            );
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

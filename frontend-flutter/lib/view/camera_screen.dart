import 'package:amcrest_viewer_flutter/locator.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../repository/cam_viewer_repository.dart';
import '../view_model/camera.viewmodel.dart';

class CameraScreen extends StatelessWidget {
  late final CamViewerRepo repo;
  late final CameraViewModel vm;

  CameraScreen({super.key, required int cameraID}) {
    repo = locator<CamViewerRepo>();
    vm = CameraViewModel(repo: repo, cameraID: cameraID);
    vm.refresh();
  }

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
        create: (context) => vm,
        child: Center(
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
                      margin: const EdgeInsets.all(10.0),
                      color: Colors.lightBlue[600],
                      child: Column(children: [
                        Text(
                          vm.title,
                          textAlign: TextAlign.center,
                          textScaleFactor: 2,
                        ),
                        vm.url != ''
                            ? Image.network(vm.url, width: 500)
                            : const Text('No image available'),
                      ]))));
        })));
  }
}

// class _CameraScreenState extends State<CameraScreen> {
//   late CameraViewModel _camViewModel;
//   final String _cameraID;

//   _CameraScreenState(String cameraID) {
//     _cameraID = cameraID;
//   }

//   @override
//   void initState() {
//     _camViewModel = Provider.of<CameraViewModel>(context, listen: false);
//     _camViewModel.cameraID = _cameraID;
//     super.initState();
//     _camViewModel.refresh();
//   }

//   @override
//   Widget build(BuildContext context) {
//     return Scaffold(
//       appBar: AppBar(
//         // Here we take the value from the CameraScreen object that was created by
//         // the App.build method, and use it to set our appbar title.
//         title: Text(widget.title),
//       ),
//       body: Center(
//         // Center is a layout widget. It takes a single child and positions it
//         // in the middle of the parent.
//         child: Consumer<CameraViewModel>(
//           builder: (context, model, child) {
//             return Center(
//                 child: Container(
//                     margin: const EdgeInsets.all(10.0),
//                     color: Colors.lightBlue[600],
//                     child: Column(children: [
//                       Text(
//                         camera.name,
//                         textAlign: TextAlign.center,
//                         textScaleFactor: 2,
//                       ),
//                       Image.network(url, width: 500),
//                     ])));
//           },
//         ),
//       ),
//     );
//   }
// }

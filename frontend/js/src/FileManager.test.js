import { FileManager } from './FileManager';
import ServiceBroker from './shared/ServiceBroker';
import { setData } from './shared/mock/file_data';
import { Time } from './time'

var files = [
    {
        "id": 7158,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7158.mp4",
        "type": 1,
        "timestamp": "2020-06-16T03:07:02Z",
        "duration_seconds": 22,
        "length": 190597
    },
    {
        "id": 7155,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7155.jpg",
        "type": 0,
        "timestamp": "2020-06-16T03:07:08Z",
        "length": 31221
    },
    {
        "id": 7156,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7156.jpg",
        "type": 0,
        "timestamp": "2020-06-16T03:07:09Z",
        "length": 32503
    },
    {
        "id": 7157,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7157.jpg",
        "type": 0,
        "timestamp": "2020-06-16T03:07:10Z",
        "length": 32585
    },
    {
        "id": 7162,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7162.mp4",
        "type": 1,
        "timestamp": "2020-06-16T05:10:50Z",
        "duration_seconds": 22,
        "length": 315149
    },
    {
        "id": 7159,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7159.jpg",
        "type": 0,
        "timestamp": "2020-06-16T05:10:57Z",
        "length": 55153
    },
    {
        "id": 7160,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7160.jpg",
        "type": 0,
        "timestamp": "2020-06-16T05:10:58Z",
        "length": 55238
    },
    {
        "id": 7161,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7161.jpg",
        "type": 0,
        "timestamp": "2020-06-16T05:10:59Z",
        "length": 55220
    },
    {
        "id": 7166,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7166.mp4",
        "type": 1,
        "timestamp": "2020-06-16T05:15:34Z",
        "duration_seconds": 22,
        "length": 312576
    },
    {
        "id": 7163,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7163.jpg",
        "type": 0,
        "timestamp": "2020-06-16T05:15:41Z",
        "length": 54120
    },
    {
        "id": 7164,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7164.jpg",
        "type": 0,
        "timestamp": "2020-06-16T05:15:42Z",
        "length": 54226
    },
    {
        "id": 7165,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7165.jpg",
        "type": 0,
        "timestamp": "2020-06-16T05:15:43Z",
        "length": 54176
    },
    {
        "id": 7170,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7170.mp4",
        "type": 1,
        "timestamp": "2020-06-16T05:28:00Z",
        "duration_seconds": 24,
        "length": 319252
    },
    {
        "id": 7167,
        "camera_id": 1,
        "path": "/api/cameras/amcrest-amcrest-1/files/7167.jpg",
        "type": 0,
        "timestamp": "2020-06-16T05:28:07Z",
        "length": 54630
    },
]





describe('FileManager', () => {

    setData(files)

    describe('Setup', () => {

        var broker = new ServiceBroker(true);

        var fm = new FileManager("test-1", broker);
       // fm.log = function () { }


        it('Should properly init time.', async () => {
            await fm.start()

            var state = fm.getState();
            var files = fm.state.files;

            var firstFile = files[0];
            var lastFile = files[files.length - 1];
            expect(state.range).toEqual({
                min: firstFile.timestamp,
                max: lastFile.timestamp,
            });

            expect(state.window).toEqual({
                start: new Time("2020-06-16T00:00:00Z"),
                end: new Time("2020-06-17T00:00Z"),
            });

            expect(state.file.id).toEqual(lastFile.id);
            expect(state.position).toEqual(lastFile.timestamp);


        });

        it('Should handle a stats update properly', async () => {

            await fm.start()

            var state = fm.getState();
            var tmin = new Time(state.range.min);
            var tmax = new Time(state.range.max);

            var min = tmin.offset(19, "second")
            var max = tmax.offset(74, "second")

            await fm._setStats({ min_date: min, max_date: max })

            state = fm.getState();

            expect(state.range.min.unix).toEqual(min.unix);
            expect(state.range.max.unix).toEqual(max.unix);
        });



        it('Should correctly box min', async () => {

            await fm.start()

            var state = fm.getState();
            var pos = new Time(state.position)


            var tmin = pos.offset(-10, "day");
            console.log(`Setting Pos: ${tmin.iso()}`)
            await fm.setPosition(tmin);

            const s = fm.getState();

            expect(s.position.date).toEqual(s.window.start.date);


        });

        it('Should correctly box max', async () => {

            await fm.start()

            var state = fm.getState();
            var pos = new Time(state.position)

            var tmax = pos.offset(10, "day");
            await fm.setPosition(tmax);

            const s = fm.getState();

            expect(s.position.unix).toEqual(s.window.end.unix);


        });

    })
});
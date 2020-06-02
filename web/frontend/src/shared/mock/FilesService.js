class FilesService {

    constructor() {

        this.files = [
            {
                "id": 860,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590843637.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:00:37Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 62924
            },
            {
                "id": 861,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590843638.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:00:38Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 63407
            },
            {
                "id": 862,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590843639.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:00:39Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 64198
            },
            {
                "id": 867,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590845538.mp4",
                "type": 1,
                "timestamp": "2020-05-30T13:32:18Z",
                "Received": null,
                "duration_seconds": 34,
                "Length": 861255
            },
            {
                "id": 864,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590845543.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:32:23Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 62321
            },
            {
                "id": 865,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590845544.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:32:24Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 62314
            },
            {
                "id": 866,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590845545.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:32:25Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 61814
            },
            {
                "id": 871,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590845567.mp4",
                "type": 1,
                "timestamp": "2020-05-30T13:32:47Z",
                "Received": null,
                "duration_seconds": 22,
                "Length": 469462
            },
            {
                "id": 868,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590845572.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:32:52Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 48826
            },
            {
                "id": 869,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590845573.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:32:53Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 46199
            },
            {
                "id": 870,
                "camera_id": 1,
                "path": "cameras/amcrest-1/1590845574.jpg",
                "type": 0,
                "timestamp": "2020-05-30T13:32:54Z",
                "Received": null,
                "duration_seconds": null,
                "Length": 53489
            }

        ];


        this.files.forEach(el => {
            var parts = el.path.split("/");
            var file = parts[parts.length - 1];
            switch (el.type) {
                case 0:
                    el.path = "/1591027084.jpg";
                    break;
                case 1:
                    el.path = "/1591028951.mp4";
                    break
            }

        });

    }

    async retrieveItems(startDate, endDate) {

        return Promise.resolve(this.files);

    }

    async getItem(id) {

        for (var i = 0; i < this.cameras.length; i++) {

            if (this.files[i].id === id) {
                return Promise.resolve(this.files[i]);
            }

        }

        return null;

    }

}

export default FilesService;
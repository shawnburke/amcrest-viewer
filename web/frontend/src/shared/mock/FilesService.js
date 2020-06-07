export var FileData = [
    {
        "id": 860,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590843637.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:00:37Z",
        "Received": null,
        "duration_seconds": null,
        "length": 62924
    },
    {
        "id": 861,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590843638.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:00:38Z",
        "Received": null,
        "duration_seconds": null,
        "length": 63407
    },
    {
        "id": 862,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590843639.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:00:39Z",
        "Received": null,
        "duration_seconds": null,
        "length": 64198
    },
    {
        "id": 867,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590845538.mp4",
        "type": 1,
        "timestamp": "2020-05-30T13:32:18Z",
        "Received": null,
        "duration_seconds": 34,
        "length": 861255
    },
    {
        "id": 864,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590845543.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:32:23Z",
        "Received": null,
        "duration_seconds": null,
        "length": 62321
    },
    {
        "id": 865,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590845544.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:32:24Z",
        "Received": null,
        "duration_seconds": null,
        "length": 62314
    },
    {
        "id": 866,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590845545.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:32:25Z",
        "Received": null,
        "duration_seconds": null,
        "length": 61814
    },
    {
        "id": 871,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590845567.mp4",
        "type": 1,
        "timestamp": "2020-05-30T13:32:47Z",
        "Received": null,
        "duration_seconds": 22,
        "length": 469462
    },
    {
        "id": 868,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590845572.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:32:52Z",
        "Received": null,
        "duration_seconds": null,
        "length": 48826
    },
    {
        "id": 869,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590845573.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:32:53Z",
        "Received": null,
        "duration_seconds": null,
        "length": 46199
    },
    {
        "id": 870,
        "camera_id": 1,
        "path": "cameras/amcrest-1/1590845574.jpg",
        "type": 0,
        "timestamp": "2020-05-30T13:32:54Z",
        "Received": null,
        "duration_seconds": null,
        "length": 53489
    }

]


FileData.forEach(f => {

    f.timestamp = new Date(f.timestamp);

})



class FilesService {

    constructor() {

        this.files = FileData;


        this.files.forEach(el => {
            switch (el.type) {
                case 0:
                    el.path = "/1591027084.jpg";
                    break;
                case 1:
                    el.path = "/1591028951.mp4";
                    break
                default:
                    break;
            }

        });

        //this.files = this.files.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());

    }

    async retrieveItems(startDate, endDate, sort) {
        console.log(`Retrieve ${startDate} => ${endDate} (${sort})`)
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


import {updateFile} from "./FilesService"
import { Time } from "../time";
import {ApiClient, Camera, DefaultApi} from "amcrest_viewer_api";


function normalizeCameraId(id) {
    if (id instanceof Camera) {
        return id.id;
    }

    if (id.toString().indexOf("-") === -1) {
        return Number(id);
    }
    var num = id.toString().match(/\d+/)[0];
    return Number(num);
}

class CamerasService {

    constructor(root) {

        this.root = root;
        this.url = (root || "") + "/api/cameras";

        this.client = new ApiClient(root);
        this.api = new DefaultApi(this.client);

    }

    updateId(cam) {
        if (!cam.intId) {
            cam.intId = normalizeCameraId(cam);
        }
        cam.intId = cam.id;
        cam.id = cam.id.toString().includes(cam.type) ? cam.id : `${cam.type}-${cam.id}`
        return cam;
    }

    async retrieveItems() {

        return new Promise((resolve, reject) => {
            this.api.getCameras({}, (error, data, response) => {
                if (error) {
                    console.error("Error retrieving cameras", error);
                    return this.handleResponseError(response);
                }
                var mapped = data.map(item => {
                    var cam = this.updateId(item);
                    if (cam.latest_snapshot) {
                        updateFile(cam.latest_snapshot, this.root)
                    }
                    return cam;
                });
                resolve(mapped);
            });
        });

      

    }

    async getItem(id) {



        return fetch(`${this.url}/${id}?latest_snapshot=1`)

            .then(response => {

                if (!response.ok) {

                    this.handleResponseError(response);

                }

                var cam = response.json();
                this.updateId(cam);
                return cam;
            })

            .then(item => {

                return item;

            }

            )

            .catch(error => {

                this.handleError(error);

            });

    }

    async getLiveStreamUrl(id) {
        return fetch(`${this.url}/${id}/live?redirect=false`)

        .then(response => {

            if (!response.ok) {

                this.handleResponseError(response);

            }

            return response.json();

        })

        .then(item => {
            if (!item || !item.uri) {
                console.log(`Didn't get uri from live request`)
            }
            if (this.root) {
                item.uri = this.root + item.uri;
            }
            return item.uri;

        }

        )

        .catch(error => {

            this.handleError(error);

        });
    }

    async getStats(id) {

        return new Promise((resolve, reject) => {
            var iid = normalizeCameraId(id);
            this.api.getCameraStats(iid, {
                start: new Date(new Date().getTime() - 1000 * 60 * 60 * 24 * 30),
                end: new Date(),
            }, 
            (error, item, response) => {
                if (error) {
                    return this.handleResponseError(response);
                }
                if (item.min_date) {
                    item.min_date = new Time(item.min_date);
                }
                if (item.max_date) {
                    item.max_date = new Time(item.max_date);
                }
                resolve(item);
            });
        });


        // return fetch(`${this.url}/${id}/stats`)

        //     .then(response => {

        //         if (!response.ok) {

        //             this.handleResponseError(response);

        //         }

        //         return response.json();

        //     })

        //     .then(item => {

        //         if (item.min_date) {
        //             item.min_date = new Time(item.min_date);
        //         }
        //         if (item.max_date) {
        //             item.max_date = new Time(item.max_date);
        //         }
        //         return item;

        //     }

        //     )

        //     .catch(error => {

        //         this.handleError(error);

        //     });

    }

    async createItem(newitem) {

        console.log("ItemService.createItem():");

        console.log(newitem);

        return fetch(this.url, {

            method: "POST",

            mode: "cors",

            headers: {

                "Content-Type": "application/json"

            },

            body: JSON.stringify(newitem)

        })

            .then(response => {

                if (!response.ok) {

                    this.handleResponseError(response);

                }

                return response.json();

            })

            .catch(error => {

                this.handleError(error);

            });

    }

    async deleteItem(itemlink) {

        console.log("ItemService.deleteItem():");

        console.log("item: " + itemlink);

        return fetch(itemlink, {

            method: "DELETE",

            mode: "cors"

        })

            .then(response => {

                if (!response.ok) {

                    this.handleResponseError(response);

                }

            })

            .catch(error => {

                this.handleError(error);

            });

    }

    async updateItem(item) {

        console.log("ItemService.updateItem():");

        console.log(item);

        return fetch(item.link, {

            method: "PUT",

            mode: "cors",

            headers: {

                "Content-Type": "application/json"

            },

            body: JSON.stringify(item)

        })

            .then(response => {

                if (!response.ok) {

                    this.handleResponseError(response);

                }

                return response.json();

            })

            .catch(error => {

                this.handleError(error);

            });

    }

    handleResponseError(response) {

        throw new Error("HTTP error, status = " + response.status + "body:" + (response && response.json && response.json()));

    }

    handleError(error) {

        console.error(error.message);

    }

}

export default CamerasService;
class FilesService {

    constructor(camera, root) {

        this.url = (root||"") + `/api/cameras/${camera}/files`;

    }

    updateTimestamp(f) {
        if (typeof f.timestamp === "string") {
            f.timestamp = new Date(f.timestamp);
        }
        return f;
    }

    async retrieveItems(startDate, endDate, sort) {

        console.log(`Fetching ${startDate.toString()} => ${endDate.toString()} (${sort})`);

        var url = `${this.url}?start=${startDate.getTime()}&end=${endDate.getTime()}&sort=${sort}`
        return fetch(url)

            .then(response => {

                if (!response.ok) {

                    this.handleResponseError(response);
                }
                return response.json();
            })

            .then(items => {
                items.forEach(f => this.updateTimestamp(f))
                return items;
            })

            .catch(error => {
                this.handleError(error);

            });

    }

    async getItem(id) {


        return fetch(`${this.url}/${id}/info`)

            .then(response => {
                if (!response.ok) {
                    this.handleResponseError(response);
                }
                return response.json();
            })
            .then(item => {
                this.updateTimestamp(item);
                return item;
            })
            .catch(error => {

                this.handleError(error);

            });

    }

    



    handleResponseError(response) {
        throw new Error("HTTP error, status = " + response.status);
    }

    handleError(error) {
        console.log(error.message);
    }
}

export default FilesService;
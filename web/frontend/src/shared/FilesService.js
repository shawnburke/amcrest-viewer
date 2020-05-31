class FilesService {

    constructor(camera) {

        this.url = `/api/cameras/${camera}/files`;

    }

    async retrieveItems(startDate, endDate) {

        var url = `${this.url}?start=${startDate.getTime()}&end=${endDate.getTime()}`
        return fetch(url)

            .then(response => {

                if (!response.ok) {

                    this.handleResponseError(response);
                }
                return response.json();
            })

            .then(json => {
                return json;
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
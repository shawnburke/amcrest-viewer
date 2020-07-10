class FilesService {

    constructor(camera, root) {

        this.root = root;
        this.url = (root||"") + `/api/cameras/${camera}/files`;

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
                items.forEach(f => updateFile(f, this.root))
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

export function updateFile(f, root) {
    if (typeof f.timestamp === "string") {
        f.timestamp = new Date(f.timestamp);
    }

    if (root) {
        f.path = root + f.path;
    }
    return f;
}


export default FilesService;
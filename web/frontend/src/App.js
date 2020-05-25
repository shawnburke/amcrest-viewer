import 'bootstrap/dist/css/bootstrap.min.css';

import React from 'react';
import './App.css';
import CameraList from "./CameraDropdown";
import CameraView from "./CameraView";
import CameraSummary from "./CameraSummary";
import { Container, Navbar, Nav } from 'react-bootstrap';

import ServiceBroker from "./shared/ServiceBroker";




import {
    HashRouter as Router,
    Route,
    useParams,
} from "react-router-dom";


class App extends React.Component {

    constructor(props) {
        super(props);

        let broker = new ServiceBroker()
        this.camsService = broker.newCamsService();

        this.state = {
            cameras: []
        }
    }

    componentDidMount() {

        this.camsService.retrieveItems().then(items => {

            this.setState({ cameras: items });

        });
    }

    render() {

        let cams = this.state.cameras || [];



        return (

            <Router>
                <div className="App">
                    <Container>
                        <Navbar bg="light" expand="lg">
                            <Navbar.Brand href="#/">Camera Viewer <Dev /></Navbar.Brand>
                            <CameraList cameras={cams} />
                            <Navbar.Toggle aria-controls="basic-navbar-nav" />
                            <Navbar.Collapse id="basic-navbar-nav">
                                <Nav className="mr-auto">
                                    <Nav.Link href="#settings">Settings</Nav.Link>
                                    <Nav.Link href="#logs">Logs</Nav.Link>

                                </Nav>

                            </Navbar.Collapse>
                        </Navbar>


                        <Route exact path="/">
                            <CameraSummary cameras={cams} />
                        </Route>
                        <Route path="/cameras/:id" component={Camera} />
                    </Container>
                </div >
            </Router >
        );
    }
}



function Camera() {
    // We can use the `useParams` hook here to access
    // the dynamic pieces of the URL.
    let { id } = useParams();

    return (
        <CameraView camera={id} />
    );
}

function Dev() {
    if (!isDev()) {
        return null;
    }
    return <span>🛠</span>;
}

function isDev() {

    return process.env.NODE_ENV !== "production";
}

export default App;

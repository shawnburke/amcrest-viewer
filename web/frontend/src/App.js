import 'bootstrap/dist/css/bootstrap.min.css';

import React from 'react';
import logo from './logo.svg';
import './App.css';
import CameraList from "./CameraDropdown";
import CameraView from "./CameraView";
import CameraSummary from "./CameraSummary";
import { Container, Button, Navbar, NavDropdown, Nav, Form, FormControl } from 'react-bootstrap';
import { Row, Col } from 'react-bootstrap';

import {
    BrowserRouter as Router,
    Switch,
    Route,
    Link,
    useParams,
} from "react-router-dom";

function App() {


    const cameras =
        [
            { name: "Garage Cam", type: "amcrest", id: "amcrest-1" },
            { name: "Front Cam", type: "amcrest", id: "amcrest-2" },
        ]

    return (
        <Router>
            <div className="App">
                <Container>
                    <Navbar bg="light" expand="lg">
                        <Navbar.Brand href="/">Camera Viewer</Navbar.Brand>
                        <CameraList cameras={cameras} />
                        <Navbar.Toggle aria-controls="basic-navbar-nav" />
                        <Navbar.Collapse id="basic-navbar-nav">
                            <Nav className="mr-auto">
                                <Nav.Link href="#settings">Settings</Nav.Link>
                                <Nav.Link href="#logs">Logs</Nav.Link>

                            </Nav>

                        </Navbar.Collapse>
                    </Navbar>

                    <Switch>
                        <Route exact path="/">
                            <CameraSummary cameras={cameras} />
                        </Route>
                        <Route path="/cameras/:id" children={<Camera />} />
                    </Switch>
                </Container>
            </div >
        </Router>
    );
}



function Camera() {
    // We can use the `useParams` hook here to access
    // the dynamic pieces of the URL.
    let { id } = useParams();

    return (
        <CameraView camera={id} />
    );
}


export default App;

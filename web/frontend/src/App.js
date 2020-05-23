import 'bootstrap/dist/css/bootstrap.min.css';

import React from 'react';
import logo from './logo.svg';
import './App.css';
import CameraList from "./CameraDropdown"
import { Container, Button, Navbar, NavDropdown, Nav, Form, FormControl } from 'react-bootstrap';
import { Row, Col } from 'react-bootstrap';

function App() {
    return (
        <div className="App">
            <Container>
                <Navbar bg="light" expand="lg">
                    <Navbar.Brand href="#home">Camera Viewer</Navbar.Brand>
                    <CameraList cameras={
                        [
                            { name: "Garage Cam", type: "amcrest", id: "amcrest-1" },
                            { name: "Frond Cam", type: "amcrest", id: "amcrest-2" },
                        ]
                    } />
                    <Navbar.Toggle aria-controls="basic-navbar-nav" />
                    <Navbar.Collapse id="basic-navbar-nav">
                        <Nav className="mr-auto">
                            <Nav.Link href="#settings">Settings</Nav.Link>
                            <Nav.Link href="#logs">Logs</Nav.Link>

                        </Nav>

                    </Navbar.Collapse>
                </Navbar>

                    <Row>
                        <Col>
                            <div style={{
                                width: "100%",
                                background: "black",
                                height: "200px",
                            }}></div>
                        </Col>
                    </Row>
                    <Row>
                        <Col></Col>
                        <Col>
                            <Button style={{ width: "100%;" }}><span>📅 Today </span></Button>
                        </Col>
                        <Col style={{ textAlign: "center" }}>
                            <Button><span>⚙</span></Button>
                        </Col>
                    </Row>
            </Container>
        </div >
            );
        }
        
        export default App;

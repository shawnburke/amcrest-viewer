import React from 'react';
import { Navbar, Nav, NavDropdown } from 'react-bootstrap';


export class Header extends React.Component  {

    constructor(props) {
        super(props);
        this.state = {
            camera: null
        }
    }

    render() {
        var camname = this.props.current && this.props.current.name;
       return  <Navbar bg="light" expand="lg">
                <Navbar.Brand href="#/">{camname || "Camera Viewer"} <Dev /></Navbar.Brand>
                <Navbar.Toggle aria-controls="basic-navbar-nav" />
                <Navbar.Collapse id="basic-navbar-nav">
                    <CameraList cameras={this.props.cams} />
                    <Nav className="mr-auto">
                        <Nav.Link href="#settings">Settings</Nav.Link>
                        <Nav.Link href="#logs">Logs</Nav.Link>
                    </Nav>
                </Navbar.Collapse>
            </Navbar>;
    };

}



class CameraList extends React.Component {
    render() {
        const rows = [];

        this.props.cameras.forEach((camera) => {
            rows.push(
                <CameraRow camera={camera} key={camera.id} />
            );
        });

        return (
            <NavDropdown title="Cameras" id="basic-nav-dropdown">
                {rows}
            </NavDropdown>
        );
    }
}


class CameraRow extends React.Component {

    render() {
        const camera = this.props.camera;
        return (
            <NavDropdown.Item href={"#cameras/" + camera.id}>{camera.name} ({camera.type})</NavDropdown.Item>
        )
    }
}



function Dev() {
    const isProd = process.env.NODE_ENV === "production";
    if (isProd) {
        return null;
    }
    return <span>🛠</span>;
}



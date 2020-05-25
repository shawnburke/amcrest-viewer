


import React from 'react';
import { NavDropdown } from 'react-bootstrap';


class CameraRow extends React.Component {

    render() {
        const camera = this.props.camera;
        return (
            <NavDropdown.Item href={"#cameras/" + camera.id}>{camera.name} ({camera.type})</NavDropdown.Item>
        )
    }
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

export default CameraList;
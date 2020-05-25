
import React from 'react';
import { Row, Col, Button } from 'react-bootstrap';


class CameraView extends React.Component {
    render() {
        document.title = "Camera Viewer - " + this.props.camera;
        return <div> <Row>
            <Col>
                <div style={{
                    width: "100%",
                    background: "black",
                    color: "white",
                    height: "200px",
                }}>{this.props.camera}</div>
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
        </div>
    }
}

export default CameraView;
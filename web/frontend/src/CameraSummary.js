
import React from 'react';
import { Row, Col, Button } from 'react-bootstrap';


class CameraSummary extends React.Component {
    render() {
        document.title = "Camera Viewer";
        var rows = this.props.cameras.map((c) => {

            return <Row style={{
                border: "thin black solid",
                background: "silver",
                padding: "5px",
                margin: "5px"

            }}>
                <Col>
                    <a href={'#cameras/' + c.id}><h1>{c.name}</h1></a>
                </Col>
            </Row>

        });

        return <div> {rows} </div>
    }
}

export default CameraSummary;
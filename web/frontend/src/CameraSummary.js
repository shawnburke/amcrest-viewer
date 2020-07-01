
import React from 'react';
import { Row, Col } from 'react-bootstrap';


class CameraSummary extends React.Component {
    render() {
        document.title = "Camera Viewer";
        var rows = this.props.cameras.map((c) => {


            var snapshot = [<span></span>];

            if (c.latest_snapshot) {
                snapshot = [<img style={{
                    maxHeight: "100px"
                }} src={c.latest_snapshot.path} />];
            }

            return <Row key={c.name} style={{
                border: "thin black solid",
                background: "silver",
                padding: "5px",
                margin: "5px"

            }}>
                <Col>
                    <a href={'#cameras/' + c.id}><h1>{c.name}</h1></a>
                    {snapshot}
                </Col>
            </Row>

        });

        return <div> {rows} </div>
    }
}

export default CameraSummary;
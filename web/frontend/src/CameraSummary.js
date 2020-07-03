
import React from 'react';
import { Row, Col } from 'react-bootstrap';


class CameraSummary extends React.Component {
    getSnapshotTime(c) {
        if (c.latest_snapshot) {
            return new Date(c.latest_snapshot.timestamp).toLocaleString();
        }
        return "";
    }
    render() {
        document.title = "Camera Viewer";
        var rows = this.props.cameras.map((c) => {


            var snapshot = [<span></span>];

            if (c.latest_snapshot) {
                snapshot = [<img alt="snapshot" style={{
                    maxWidth: "100%",
                    border: "thin solid black",
                }} src={c.latest_snapshot.path} />];
            }

            const radius = "5px";
            const boxShadow = "2px 2px 8px #888888";

            return <Row key={c.name} style={{
                border: "thin black solid",
                background: "#DDD",
                padding: "5px",
                margin: "5px",
                MozBorderRadius: radius,
                WebkitBorderRadius: radius,
                MozBoxShadow: boxShadow,
                WebkitBoxShadow: boxShadow,
                boxShadow: boxShadow,
            }}>
                <Col>
                    <a href={'#cameras/' + c.id}>
                        {snapshot}
                        <Row>
                            <Col>
                                <div style={{ textAlign: "left" }}>
                                    <h4>{c.name}</h4>
                                </div>
                            </Col>
                            <Col>
                                <div style={{ textAlign: "right", fontSize: ".75em" }}>
                                    {this.getSnapshotTime(c)}
                                </div>
                            </Col>
                        </Row>

                    </a>
                </Col>
            </Row>

        });

        return <div> {rows} </div>
    }
}

export default CameraSummary;
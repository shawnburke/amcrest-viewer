
import React from 'react';
import { Row, Col } from 'react-bootstrap';


class CameraSummary extends React.Component {
    getSnapshotTime(c) {
        if (c.latest_snapshot) {
            var d = new Date(c.latest_snapshot.timestamp);
            return {
                label:d.toLocaleString(),
                seconds_old: Math.trunc((new Date().getTime() - d.getTime())/ 1000),
            } ;
        }
        return {};
    }
    componentDidUpdate() {
        document.title = "Camera Viewer";
    }
    render() {
        
        var rows = this.props.cameras.map((c) => {


            var snapshot = [<span key='span0'></span>];

            if (c.latest_snapshot) {
                snapshot = [<img  key='snapshot-0' alt="snapshot" style={{
                    maxWidth: "100%",
                    border: "thin solid black",
                }} src={c.latest_snapshot.path} />];
            }

            const radius = "5px";
            const boxShadow = "2px 2px 8px #888888";
            var snapshotTime = this.getSnapshotTime(c)

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
                                    {snapshotTime.label}<br/>
                                    {snapshotTime.seconds_old}s
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
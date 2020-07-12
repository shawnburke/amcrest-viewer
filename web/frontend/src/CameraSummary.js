
import React from 'react';
import { Row, Col } from 'react-bootstrap';
import {Time, second} from "./time";


class CameraSummary extends React.Component {
    
    componentDidUpdate() {
        document.title = "Camera Viewer";
    }
    render() {
        
        var rows = this.props.cameras.map((c) => {


          return <CamaraSummaryItem key={"camera-" + c.id} camera={c}/>

        });

        return <div> {rows} </div>
    }
}

class CamaraSummaryItem extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            snapshot_time: null,
        }

    }


    componentDidMount() {
        const c = this.props.camera;

        if (!c) {
            return;
        }

        if (c.latest_snapshot) {
            var t = c.latest_snapshot.timestamp;
            this.setState({
                time_label: t.date.toLocaleString(),
                since_snapshot: Time.now().delta(t, second),
                timestamp: t.round(second),
            });
            setTimeout(this.ticker.bind(this), 1000);
        }

    }

    componentWillUnmount() {
        this.done = true;
    }

    ticker() {

        if (this.done) {
            return;
        }

        
        var since = Time.now().delta(this.state.timestamp, second);

        this.setState({
            since_snapshot: since,
        });

        setTimeout(this.ticker.bind(this), 1000);

    }


    render() {

        const c = this.props.camera;

        if (!c) {
            return <div>&nbsp;</div>
        }

        var snapshot = [<span key='span0'></span>];

        if (c.latest_snapshot) {
            snapshot = [<img  key='snapshot-0' alt="snapshot" style={{
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
                                {this.state.time_label}<br/>
                                {this.state.since_snapshot}s
                            </div>
                            
                        </Col>
                    </Row>

                </a>
            </Col>
        </Row>

    }
}

export default CameraSummary;
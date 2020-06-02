
import React from 'react';
import { Row, Col, Button } from 'react-bootstrap';
// import { DatePicker } from 'react-datepicker';
//import "react-datepicker/dist/react-datepicker.css";
import DatePicker from 'react-date-picker'
import ServiceBroker from "./shared/ServiceBroker";
import ReactPlayer from 'react-player'


class CameraView extends React.Component {
    constructor(props) {
        super(props);

        let broker = new ServiceBroker()

        // TODO: fix hack
        this.filesService = broker.newFilesService("amcrest-" + this.props.camera);

        this.state = {
            files: [],
            date: new Date(),
        }
    }

    componentDidMount() {

        this.loadFiles();
    }

    loadFiles() {
        var date = this.state.date;
        var start = date.toString().replace(/\d{2}:\d{2}:\d{2}/, "00:00:00")
        start = new Date(start);

        var end = new Date(start.getTime() + (24 * 60 * 60 * 1000));


        this.filesService.retrieveItems(start, end).then(items => {

            this.setState({ files: items, source: null });

        });
    }


    handleClick(el) {
        el.preventDefault();
        this.setState(
            {
                source: el.target.attributes.src
            }
        );
    }


    render() {

        document.title = "Camera Viewer - " + this.props.camera;

        var fileRows = [];

        if (this.state.files) {
            this.state.files.forEach(f => {
                var row = <Row key={f.id}>
                    <Col>{new Date(f.timestamp).toLocaleTimeString()}</Col>
                    <Col>{f.type}</Col>
                    <Col>{f.duration_seconds}</Col>
                    <Col><a src={f.path} onClick={this.handleClick.bind(this)}>{f.path}</a></Col>
                </Row>;

                fileRows.push(row);
            })
        }

        return <div> <Row>
            <Col>
                <div style={{
                    width: "100%",
                    background: "black",
                    color: "white",
                    height: "200px",
                }}><Player source={this.state.source} /></div>
            </Col>
        </Row>
            <Row>
                <Col></Col>
                <Col>
                    <DatePicker
                        todayButton="Today"
                        value={this.state.date}
                        onChange={date => this.setStartDate(date)}
                    />
                </Col>
                <Col style={{ textAlign: "center" }}>
                    <Button><span>⚙</span></Button>
                </Col>
            </Row>
            {fileRows}
        </div>
    }

    setStartDate(d) {

        if (d === this.state.date) {
            return;
        }
        this.setState({
            date: d,
        })
        this.loadFiles();
    }
}




class Player extends React.Component {

    render() {

        if (this.props.source == null) {
            return <div></div>;
        }


        var val = this.props.source.value;
        var mp4 = val.endsWith("mp4");

        if (mp4) {
            return <ReactPlayer url={val} width="100%" height="100%" playing="true" style={{
                height:"100%"
            }} />;
        }
        var jpg = val.endsWith("jpg");

        if (jpg) {
            return <img src={val}  style={{
                height: "100%",
            }}/>;
        }

        return <div></div>;

    }
}

export default CameraView;

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
            selected: 0
        }
    }

    componentDidMount() {

        this.loadFiles();
    }

    componentDidUpdate(prevProps, prevState) {  
        if (this.state.date !== prevState.date) {    
            this.loadFiles(this.state.date);
        }  
    }

    loadFiles(date) {
        date = date || this.state.date;
        var start = date.toString().replace(/\d{2}:\d{2}:\d{2}/, "00:00:00")
        start = new Date(start);

        var end = new Date(start.getTime() + (24 * 60 * 60 * 1000)-1);


        this.filesService.retrieveItems(start, end, "desc").then(items => {

            this.setState({ 
                files: items, 
                source: null, 
                selected:0 
            });

        });
    }


    handleClick(el) {
        el.preventDefault();
        var index = Number(el.currentTarget.attributes.file.value);
        this.setState(
            {
                source: this.state.files[index],
                selected: index
            }
        );
    }


    render() {

        document.title = "Camera Viewer - " + this.props.camera;

        var fileRows = [];

        if (this.state.files) {
            this.state.files.forEach((f,i) => {

                var style = {};

                if (i === this.state.selected) {
                    style = {
                        background: "yellow"
                    }
                }

                var t = "jpg";

                if (f.type === 1) {
                    t = "mp4";
                }

                var row = <Row key={f.id} style={style} file={i} onClick={this.handleClick.bind(this)}>
                    <Col >{new Date(f.timestamp).toTimeString()}</Col>
                    <Col>{t}</Col>
                    <Col>{f.duration_seconds}</Col>
                    <Col><a href="{f.path}" target="_vid">{f.path}</a></Col>
                </Row>;

             
                fileRows.push(row);
            })
        }

        var windowHeight = window.innerHeight;

        return <div> <Row>
            <Col>
                <div style={{
                    width: "100%",
                    background: "black",
                    color: "white",
                    height: windowHeight * .4,
                }}><Player file={this.state.source} /></div>
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
            <div style={{
                maxHeight:  windowHeight * .5,
                overflowY: "auto",
                overflowX: "hidden"
            }}>
            {fileRows}
            </div>
        </div>
    }

    setStartDate(d) {

        if (d === this.state.date) {
            return;
        }
        this.setState({
            date: d,
        })
    }
}




class Player extends React.Component {

    render() {

        if (this.props.file == null) {
            return <div></div>;
        }


        var val = this.props.file;
        

        if (val.type === 1) {
            return <ReactPlayer 
                controls
                url={val.path} 
                width="100%" 
                height="100%" 
                playsinline="true"
                playing="true" style={{
                    height:"100%"
                }} 
            />;
        }
        
        if (val.type === 0) {
            return <img alt="view" src={val.path}  style={{
                height: "100%"
            }}/>;
        }

        return <div></div>;

    }
}

export default CameraView;
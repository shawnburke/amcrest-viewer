
import React from 'react';

import { Row, Col, Button } from 'react-bootstrap';

import ServiceBroker from "./shared/ServiceBroker";
import Player from "./Player";

import DatePicker from "./DatePicker";
import TimeScroll from "./TimeScroll";
import { day, second, Time } from './time';
import { FileManager } from "./FileManager";
import { FileList } from "./FileList"


class CameraView extends React.Component {
    constructor(props) {
        super(props);

        // TODO: fix hack
        this.camid = this.props.cameraid;

        if (this.camid.toString().indexOf("-") === -1) {
            this.camid = "amcrest-" + this.camid;
        }

        this.serviceBroker = new ServiceBroker();

        this.fileManager = new FileManager(this.camid, this.serviceBroker);


        this.fileManager.onChange = this._fileManagerChange.bind(this);

        this.state = {
            files: [],
            date: new Time(),
            position: new Time(),
            selected: 0,
            range: {
                min: new Time("2000-01-01"),
                max: new Time(),
            },
            window: {
                start: new Time().offset(-1, day),
                end: new Time()
            },
            mediaItems: [],
        }
    }

    componentDidMount() {

        this.fileManager.start();

    }


    _fileManagerChange(change) {

        if (change.range) {
            this.setState({
                range: change.range
            })
        }

        if (change.window) {

            var d = change.window.end.floor(day);

            this.setState({
                window: change.window,
                date: d,
            })
        }

        if (change.files) {
            var items = this.getTimeItems(change.files);
            this.setState({
                mediaItems: items
            })
        }



        if (change.source !== undefined) {
            this.setState({
                source: change.source
            })
        }



        if (change.file !== undefined) {
            this.setState({
                source: change.file
            })
        }

        if (change.position !== undefined) {
            this.setState({
                position: change.position
            })
        }


    }

    isLiveView() {
        var isLive = this.state.source && this.state.source.type === 2;
        return isLive;
    }

    onTimeScrollChange(time, item) {

        this.fileManager.setPosition(time, item && item.file);

    }

    onSelectedFileChange(f) {

        this.fileManager.setCurrentFile(f);

    }

    stopLiveView() {
        this.fileManager.stopLive();
    }

    onLiveClick(ev) {

        if (this.isLiveView()) {
            this.stopLiveView();
            return;
        }

        var target = ev.currentTarget;

        target.disabled = true;
        this.fileManager.startLive().then(_success => {
            target.disabled = false;
        })
    }

    getTimeItems(files) {
        if (!files) {
            return []
        }

        const jpgSeconds = 60;

        var items = files.map(f => {

            var sec = f.duration_seconds || jpgSeconds;
            var end = f.timestamp.offset(sec, second);
        
            return {
                id: f.id,
                start: f.timestamp,
                end: end,
                video: f.type === 1,
                source: f.path,
                file: f,
            }
        })

        return items;
    }


    render() {

        document.title = "Camera Viewer - " + this.props.cameraid;
        const secondaryBackground = "#EEE";
        var windowHeight = window.innerHeight;
        var pos = "";

        if (this.state.position) {
            pos = this.state.position.date.toLocaleTimeString();
        }

        return <div style={{}}>
            <Row>
                <Col style={{
                    position: "relative",
                    background: "black",
                    color: "white",
                    height: windowHeight * .4,
                    borderRadius: "5px 5px 0px 0px",
                }}>
                    <Player file={this.state.source} position={this.state.position} />
                </Col>
            </Row>
            <Row style={{ background: secondaryBackground }}>

                <Col xs={8}>
                    <DatePicker
                        minDate={this.state.range.min}
                        maxDate={this.state.range.max}
                        date={this.state.date}
                        onChange={date => this.setCurrentDate(date)}
                    />
                </Col>
                <Col xs={4} style={{ textAlign: "right" }}>
                    {pos}
                </Col>

            </Row>
            <Row>
                <Col xs={10} style={{ margin: "2px", background: "black" }}>

                    <TimeScroll
                        startTime={this.state.window.start}
                        endTime={this.state.window.end}
                        items={this.state.mediaItems}
                        position={this.state.position}
                        onTimeChange={this.onTimeScrollChange.bind(this)}
                    />

                </Col>
                <Col xs={1} ><Button style={{
                    marginTop: "20px",
                    background: this.state.source && this.state.source.type === 2 ? "red" : "blue",
                }}
                    onClick={this.onLiveClick.bind(this)}>Live</Button></Col>
            </Row>
            <Row>
                <Col style={{ background: secondaryBackground }}>
                    <FileList
                        selected={this.state.source}
                        onSelectedFileChange={this.onSelectedFileChange.bind(this)}
                        files={this.fileManager.files}
                    />
                </Col>
            </Row>
        </div >

    }

    setCurrentDate(d) {

        if (this.state.date.same(d)) {
            return;
        }

        d = d.floor(day);
        var e = d.add(1, day);
       
        this.fileManager.setWindow(d, e);
    }
}

export default CameraView;
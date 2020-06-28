
import React from 'react';

import { Row, Col } from 'react-bootstrap';

import ServiceBroker from "./shared/ServiceBroker";
import Player from "./Player";

import DatePicker from "./DatePicker";
import TimeScroll from "./TimeScroll";
import { day, toUnix } from './time';
import { FileManager } from "./FileManager";
import { FileList } from "./FileList"




class CameraView extends React.Component {
    constructor(props) {
        super(props);

        // TODO: fix hack
        var camid = "amcrest-" + this.props.cameraid;

        this.fileManager = new FileManager(camid, new ServiceBroker());


        this.fileManager.onChange = this._fileManagerChange.bind(this);

        this.state = {
            files: [],
            date: new Date(),
            position: new Date(),
            selected: 0,
            range: {
                min: new Date(2000, 1, 1),
                max: new Date(),
            },
            window: {
                start: new Date(new Date().getTime() - day),
                end: new Date()
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

            var d = this.fileManager.snapTime(change.window.end, "day", -1);

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

        if (change.file) {
            this.setState({
                source: change.file
            })
        }

        if (change.position) {
            this.setState({
                position: change.position
            })
        }
    }

    onTimeScrollChange(time, item) {

        this.fileManager.setPosition(time, item && item.file);
    }

    onSelectedFileChange(f) {

        this.fileManager.setCurrentFile(f);

    }

    getTimeItems(files) {
        if (!files) {
            return []
        }

        var items = files.map(f => {

            var sec = f.duration_seconds || 5;

            var end = new Date(toUnix(f.timestamp) + (1000 * sec));

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

        var windowHeight = window.innerHeight;

        return <div style={{ background: "black", color: "white" }}> <Row>
            <Col>
                <div style={{
                    width: "100%",
                    background: "black",
                    color: "white",
                    height: windowHeight * .4,
                }}><Player file={this.state.source} position={this.state.position} /></div>
            </Col>
        </Row>
            <Row>

                <Col xs={9}>
                    <DatePicker
                        minDate={this.state.range.min}
                        maxDate={this.state.range.max}
                        date={this.state.date}
                        onChange={date => this.setDate(date)}
                    />
                </Col>
                <Col xs={3}>
                    <span>{this.state.position.toLocaleTimeString()}</span>
                </Col>

            </Row>
            <div style={{ margin: "2px" }}>
                <TimeScroll
                    startTime={this.state.window.start}
                    endTime={this.state.window.end}
                    items={this.state.mediaItems}
                    position={this.state.position}
                    onTimeChange={this.onTimeScrollChange.bind(this)}
                />
            </div>

            <FileList
                selected={this.state.source}
                onSelectedFileChange={this.onSelectedFileChange.bind(this)}
                files={this.fileManager.files}
            />

        </div>
    }

    setDate(d) {

        if (d === this.state.date) {
            return;
        }

        d = this.fileManager.snapTime(d, "day", -1);
        var e = this.fileManager.dateAdd(d, 1, "day");
        e = new Date(new Date(e).getTime() - 1);

        this.fileManager.setWindow(d, e);
    }
}

export default CameraView;
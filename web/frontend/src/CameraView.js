
import React from 'react';

import { Row, Col } from 'react-bootstrap';

import ServiceBroker from "./shared/ServiceBroker";
import Player from "./Player";

import DatePicker from "./DatePicker";
import TimeScroll from "./TimeScroll";
import { day, toUnix } from './time';
import { FileManager } from "./FileManager";




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

    mediaRowClick(f, el) {
        el.preventDefault();

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


    renderFileList(files) {
        var fileRows = [];

        if (files) {


            var curmp4;

            var grouped = [];


            function finish() {

                curmp4 = null;
            }

            function group(f) {
                if (f.type !== 0) {
                    return false;
                }
                if (curmp4 && f.timestamp < curmp4.end) {
                    curmp4.children = curmp4.children || [];
                    curmp4.children.push(f);
                    return true;
                }
                return false;
            }


            // group files

            files.forEach((f) => {

                if (group(f)) {
                    return;
                }

                if (f.type === 1) {
                    finish();
                    curmp4 = f;
                    f.end = new Date(f.timestamp.getTime() + (1000 * f.duration_seconds));
                }
                f.children = null;
                grouped.push(f);
            });


            // walk through the grouped files and create rows
            //
            grouped.forEach(f => {
                var row = <FileRow
                    file={f} key={f.id}
                    selected={this.state.source && this.state.source.id === f.id}
                    onClick={this.mediaRowClick.bind(this, f)} />;

                fileRows.push(row);
            })


            fileRows = fileRows.reverse();

        }
        return fileRows;
    }


    render() {

        document.title = "Camera Viewer - " + this.props.cameraid;

        var windowHeight = window.innerHeight;

        return <div> <Row>
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

                <Col xs={12}>
                    <DatePicker
                        minDate={this.state.range.min}
                        maxDate={this.state.range.max}
                        date={this.state.date}
                        onChange={date => this.setDate(date)}
                    />
                    <span>{this.state.position.toString()}</span>
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
            <div style={{
                maxHeight: windowHeight * .5,
                overflowY: "auto",
                overflowX: "hidden"
            }}>
                {this.renderFileList(this.fileManager.files)}
            </div>
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

class FileRow extends React.Component {


    last(array) {
        if (!array || !array.length) {
            return null;
        }
        return array[array.length - 1];
    }

    render() {
        var style = {};
        const f = this.props.file;

        if (this.props.selected) {
            style = {
                background: "yellow"
            }
        }

        var t = "";

        if (f.type === 1) {
            t = "🎥";
        }

        var children = null;




        var rows = [<Row key={f.id} style={style} file={f} >
            <Col xs={1}><span role="img">{t}</span></Col>
            <Col xs={4}>{f.timestamp.toLocaleTimeString()}</Col>
            <Col xs={1}>{f.duration_seconds}</Col>
            <Col ><a href={f.path} target="_vid">{this.last(f.path.split('/'))}</a></Col>
        </Row>];

        if (f.children) {
            children = f.children.map(fc => {
                return <img key={"thumb-" + fc.id} alt="thumb" src={fc.path} file={fc.id} style={{
                    width: "40px",
                    marginLeft: "2px",
                }} />;
            })
            rows.push(<Row>
                <Col xs={1}><span></span></Col>
                <Col>{children}</Col>
                <Col xs={1}></Col>
            </Row>);
        }

        return <div onClick={this.props.onClick}>{rows}</div>;


    }
}


export default CameraView;
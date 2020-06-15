
import React from 'react';
import { Row, Col } from 'react-bootstrap';

import ServiceBroker from "./shared/ServiceBroker";
import ReactPlayer from 'react-player';

import DatePicker from "./DatePicker";
import TimeScroll from "./TimeScroll";



const hour = 3600 * 1000;
const day = hour * 24;
const month = day * 30;


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
                min:new Date(2000,1,1),
                max:new Date(),
            },
            window:{
                start:new Date(new Date().getTime() - day),
                end: new Date()
            },
            mediaItems:[],
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
            this.setState({
                window: change.window,
                date: change.window.start,
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

    onTimeScrollChange(time, item_id) {
      
        this.fileManager.setPosition(time)
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

            var end = new Date(f.timestamp.getTime() + (1000*sec));

            return {
                id: f.id,
                start: f.timestamp,
                end: end,
                video: f.type === 1,
                source: f.path,
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

                if (f.type === 1){
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
                    selected={this.state.selected === f.id}
                    onClick={this.mediaRowClick.bind(this, f)}/>;

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
                }}><Player file={this.state.source} /></div>
            </Col>
        </Row>
            <Row>
              
                <Col xs={12}>
                    <DatePicker
                        minDate={this.state.range.min}
                        maxDate={this.state.range.max}
                        date={this.state.date}
                        onChange={date => this.setStartDate(date)}
                    />
                </Col>
               
            </Row>
            <div style={{margin:"2px"}}>
            <TimeScroll 
                startTime={this.state.window.start} 
                endTime={this.state.window.end}
                items={this.state.mediaItems}
                position={this.state.date}
                onTimeChange={this.onTimeScrollChange.bind(this)}
            />       
            </div>
            <div style={{
                maxHeight:  windowHeight * .5,
                overflowY: "auto",
                overflowX: "hidden"
            }}>
            {this.renderFileList(this.fileManager.files)}
            </div>
        </div>
    }

    setStartDate(d) {

        if (d === this.state.date) {
            return;
        }
        this.fileManager.setWindow(d);
        // this.setState({
        //     date: d,
        // })
    }
}

class FileRow extends React.Component {
  

    last(array) {
        if (!array || !array.length) {
            return null;
        }
        return array[array.length-1];
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
                return <img src={fc.path} file={fc.id} style={{
                    width:"40px",
                    marginLeft:"2px",
                }}/>;
            })
            rows.push( <Row>
                <Col xs={1}><span></span></Col>
                <Col>{children}</Col>
                <Col xs={1}></Col>
            </Row>);
        }

       return <div onClick={this.props.onClick}>{rows}</div>;
      

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
                playsinline={true}
                playing={true} 
                style={{
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


// FileManager takes in a list of files
// and manages them by time
//
// Range: 
//      Min: minimum date for available files
//      Max: maximum date for available files
//
// Window: Range.Min < Window < Range.Max
//      Start/End: Time range of files currently 
//                 in scope, e.g. a given day
//      Change: Sets current to media file closest to 
//              current time
//
// Position: Window.Start < Position < Window.End
//      Change: selects a file
//              If outside window, may bump the window?
//
// Current: File that contains the Postition time.  Setting this sets Position to beginning of file
//
// Events:
//      OnCurrentFileChange
//      OnWindowChange
//      OnPositionChange
// 
class FileManager {

    constructor(camid, broker) {
        this.camid = camid;
        this.filesService = broker.newFilesService(camid);
        this.camerasServer = broker.newCamsService();

        // initialize info
        var today = new Date();
        this.range = {
            min: this.dateAdd(today, -1, "m"),
            max: today,
        }

        this.window = {
            start: this.dateAdd(today, -1, "d"),
            end: today,
        }
        
        this.position = this.dateAdd(today, -1, "h");
    }

    start() {
        this.camerasServer.getStats(this.camid).then(
            s => {
                this.setRange(new Date(s.min_date),  new Date(s.max_date));
            }
        );
    }

   
    loadFiles(start, end) {
        
        start = start.toString().replace(/\d{2}:\d{2}:\d{2}/, "00:00:00")
        start = new Date(start);

        end = end || new Date(start.getTime() + (24 * 60 * 60 * 1000)-1);


        return this.filesService.retrieveItems(start, end, "");
    }

    _onchange(value) {

        if (!value){
            return;
        }

        var rangeChange = value.range;
        if (rangeChange) {
            console.log(`Changing range to ${value.range.min} => ${value.range.max}`);
        }
        var windowChange = value.window;
        if (windowChange) {
            console.log(`Changing window to ${value.window.start} => ${value.window.end}`)
        }
        var positionChange = value.position;
        if (positionChange) {
            console.log(`Setting position to ${value.position}`);
        }

        var fileChange = value.file;
        if (fileChange) {
            console.log(`Setting file to ${value.file.id} ${value.file.path}`)
        }

        Object.assign(this, value);

        if (this.onChange) {
            this.onChange(value);
        }

    }

    snapTime(t, unit, bias) {

        var wasDate = false;
        if (t.getTime) {
            t = t.getTime();
            wasDate = true;
        }

        var chunk = 0;

        
        var offset = 0;

        switch (unit) {
            case "hour":
                chunk = hour;
                break;
            case "day":
                chunk = day;
                offset = new Date().getTimezoneOffset() * 60 * 1000;
                break;
            default:
                throw new Error("Unknown unit: " + unit); 
        }

        var delta = t % chunk;

        if (!bias) {
            bias = (delta < chunk / 2) ? -1 : 1;
        }

        switch (bias) {
            case -1:
                t -= delta;
                break;
            case 1:
                t += (chunk-delta);
                break;
        }

        t += offset;

        if (wasDate) {
            t = new Date(t);
        }
        return t;
    }

    boxTime(t, min, max, bias) {
        var toTime = function(date) {
            if (date.constructor.name === "Date") {
                return date.getTime();
            }
            return date;
        }

        var tt = toTime(t);
        var wasDate = tt !== t;
        var tmin = toTime(min);
        var tmax = toTime(max);

        var unboxed = tt < tmin || tt > tmax;

        if (unboxed) {

            switch (bias) {
                case "min":
                    tt = tmin;
                    break;
                case "max":
                    tt = tmax;
                    break;
                default:
                    if (tt < tmin) {
                        tt = tmin;
                    } else if (tt > tmax) {
                        tt = tmax;
                    }
            }
        }

        if (wasDate) {
            return new Date(tt);
        }
        return tt;
    }

    setRange(min, max) {
        if (max < min) {
            throw new Error("bad range");
        }

        if (min === this.range.min && max === this.range.max) {
            return;
        }

        this._onchange({range:{
            min: min,
            max: max,
        }
        });
        
        this.setWindow(this.window.start, this.window.end);
    }

    setWindow(start, end) {

        if (end < start) {
            throw new Error("bad window");
        }

        var boxedStart = this.boxTime(start, this.range.min, this.range.max, "min");
        var boxedEnd = this.boxTime(end, this.range.min, this.range.max, "max");

        boxedStart = this.snapTime(boxedStart, "day", -1);
        boxedEnd = this.snapTime(boxedEnd, "day", 1);

        if (boxedStart === this.window.start && boxedEnd === this.window.end) {
            return true;
        }

        this._onchange({
            window:{
            start:boxedStart,
            end:boxedEnd,
        }})
        ;

        console.log(`Loading files for range ${boxedStart} => ${boxedEnd}`)
        this.loadFiles(boxedStart, boxedEnd).then(items =>{

            console.log(`Loaded ${items.length} files`);
            var files = items.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());

            // set an end for each file item;
            files.forEach(file => {
                if (!file.end) {
                    var end = file.timestamp.getTime();

                    if (file.duration_seconds){
                        end += 1000*file.duration_seconds;
                    } else {
                        end += 1000;
                    }
                    file.start = file.timestamp;
                    file.end = end;
                }
            });

            this._onchange({files:files});

            var pos = this.boxTime(this.position, boxedStart, boxedEnd);

            // find the last file
            if (pos !== this.position && this.files.length) {
                var lastFile = this.files[this.files.length-1];

                pos = lastFile.timestamp;
            }

            this.setPosition(pos);
        })
    }

    isInFile(time, file) {

        return time >= file.start && time < file.end;
    }

    setPosition(time) {

        var boxed = this.boxTime(time, this.window.start, this.window.end, 'min');

        if (this.timeEqual(boxed, this.position)){
            return true;
        }

       
        this._onchange({position:boxed});


        // find the file
        var file = this.files.find(f => this.isInFile(time, f));

        if (file) {
            this.setCurrentFile(file);
        }

        
    }

    timeEqual(t1, t2) {
        if (t1 === t2) {
            return true;
        }

        if (!t1 || !t2) {
            return false;
        }

        
        if (t1.getTime) {
            t1 = t1.getTime()
        }

        if (t2.getTime) {
            t2 = t2.getTime()
        }

        return t1 === t2;
    }

    setCurrentFile(file) {
        var oldid = this.file && this.file.id;
        var newid = file && file.id;
        if (oldid === newid) {
            return true;
        }

        file = this.files.find(f => f.id === newid);

        if (!file) {
            console.warn(`Can't find file ${newid}`);
            return false;
        }

        // set position if not in file
        var boxed = this.boxTime(file.timestamp, this.window.start, this.window.end);

        if (!this.timeEqual(boxed,file.timestamp)) {
            console.warn(`Selected file timestamp ${file.timestamp} outside of window ${this.window.start} => ${this.window.end}`);
            return false;
        }

        this.setPosition(file.timestamp);

        this._onchange({file:file});
        return true;

    }

    dateAdd(date, n, unit) {

      
        var base = 0;
        
        switch (unit) {
            case "h": 
                base = hour;
                break;
            case "d":
                base = day;
                break;
            case "m":
                base = month;
                break;
            default:
                throw new Error("Unknown unit: " + unit);
        }

        base *= n;

        if (!date) {
            date = new Date();
        }

        return new Date(date.getTime() + base);
    }

}

export default CameraView;
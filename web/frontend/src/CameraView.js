
import React from 'react';
import { Row, Col } from 'react-bootstrap';

import ServiceBroker from "./shared/ServiceBroker";
import ReactPlayer from 'react-player';

import DatePicker from "./DatePicker";
import TimeScroll from "./TimeScroll";


class CameraView extends React.Component {
    constructor(props) {
        super(props);

        let broker = new ServiceBroker();

        // TODO: fix hack
        this.filesService = broker.newFilesService("amcrest-" + this.props.cameraid);
        this.camerasServer = broker.newCamsService();

        this.state = {
            files: [],
            date: new Date(),
            selected: 0,
            minDate: new Date(2000,1,1),
            maxDate: new Date()
        }
    }

    componentDidMount() {

       
        this.camerasServer.getStats(this.props.cameraid).then(
            s => {
                var dates = {
                    minDate: new Date(s.min_date),
                    maxDate: new Date(s.max_date)
                }
                this.setState(dates)
                this.loadFiles();

            }
        )
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


        this.filesService.retrieveItems(start, end, "").then(items => {

            this.setState({ 
                files: items, 
                source: null, 
                selected:0 
            });

        });
    }


    handleClick(f, el) {
        el.preventDefault();

        var id = f.id;

        if (el.target.attributes.file) {
            f = this.state.files.find(file => file.id == el.target.attributes.file.value);
        }
        
        this.setState(
            {
                source:f,
                selected: id
            }
        );
    }

  

    render() {

        document.title = "Camera Viewer - " + this.props.cameraid;

        var fileRows = [];

        if (this.state.files) {
            

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

            this.state.files.forEach((f) => {
                
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
                    onClick={this.handleClick.bind(this, f)}/>;

                fileRows.push(row);
            })

           
            fileRows = fileRows.reverse();

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
              
                <Col xs={12}>
                    <DatePicker
                        minDate={this.state.minDate}
                        maxDate={this.state.maxDate}
                        date={this.state.date}
                        onChange={date => this.setStartDate(date)}
                    />
                </Col>
                {/* <Col xs={1} style={{ textAlign: "center" }}>
                    <Button><span>⚙</span></Button>
                </Col> */}
            </Row>
            <TimeScroll startTime={new Date(new Date().getTime() - (60*60*1000*24))} endTime={new Date()}/>       
          
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
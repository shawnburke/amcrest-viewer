

import React from 'react';

import { Row, Col } from 'react-bootstrap';

//import { day, toUnix } from './time';


export class FileList extends React.Component {

    constructor(props) {
        super(props);
        this.myRef = React.createRef();

    }

    mediaRowClick(f, el) {
        el.preventDefault();

        if (this.props.onSelectedFileChange) {
            this.props.onSelectedFileChange(f);
        }
    }



    componentDidUpdate() {

        var selectedId = this.props.selected && this.props.selected.id;

        if (!selectedId) {
            return;
        }

        var selectedRowEl = document.getElementById(this.getRowId(selectedId));

        if (!selectedRowEl) {
            return;
        }


        var parent = this.myRef.current || selectedRowEl.parentElement;


        var top = selectedRowEl.offsetTop - parent.offsetTop;

        parent.scrollTop = top;
    }

    getRowId(id) {
        return `filerow-${id}`;
    }

    render() {
        var files = this.props.files;

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
                    if (curmp4.children.length === 1) {
                        curmp4.image = f.path;
                    }
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

                if (f.type !== 1) {
                    return;
                }

                var row = <FileRow id={this.getRowId(f.id)}
                    file={f} key={f.id}
                    className="filerow"
                    selected={this.props.selected && this.props.selected.id === f.id}
                    onClick={this.mediaRowClick.bind(this, f)} />;

                fileRows.push(row);
            })


            fileRows = fileRows.reverse();

        }
        var windowHeight = window.innerHeight;

        return <div ref={this.myRef.current} style={{
            maxHeight: windowHeight * .5,
            overflowY: "auto",
            overflowX: "hidden"
        }}>{fileRows}</div>
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
        var style = {
            paddingBottom: "2px"
        };


        const f = this.props.file;

        if (this.props.selected) {
            style.background = "lightskyblue";
        }


        return <Row
            id={this.props.id} onClick={this.props.onClick} key={'filerow' - f.id} style={style} file={f} >
            <Col xs={3} style={{ textAlign: "left" }}>
                <img src={f.image} alt="thumb" style={{
                    height: "70px",
                    border: 'thin solid black'
                }} />
            </Col>
            <Col xs={8} style={{ textAlign: "left" }}>
                <div>{f.timestamp.toLocaleTimeString()}</div>

                <div>{f.duration_seconds}s</div>
            </Col>
        </Row>;






    }
}


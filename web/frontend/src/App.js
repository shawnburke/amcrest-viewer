import 'bootstrap/dist/css/bootstrap.min.css';

import React from 'react';
import './App.css';

import CameraView from "./CameraView";
import CameraSummary from "./CameraSummary";
import { Container} from 'react-bootstrap';
import {Header} from './Header';
import ServiceBroker from "./shared/ServiceBroker";




import {
    HashRouter as Router,
    Route,
    useParams,
} from "react-router-dom";


class App extends React.Component {

    constructor(props) {
        super(props);

        


        let broker = new ServiceBroker()
        this.camsService = broker.newCamsService();

        this.state = {
            cameras: [],
            current: null,
        }
    }

    componentDidMount() {

       if (this.state.cameras.length === 0){

            this.camsService.retrieveItems().then(items => {

                var state = {
                    cameras: items 
                }
    
                this.setState(state);

            });
        }
    }

    onShowCamera(camid) {
        var cc = this.state.cameras.find(cam => cam.id === camid);

        if (cc) {
            this.setState({current:cc});
        }
    }

    render() {

        let cams = this.state.cameras || [];


        return (

            <Router>
                <div className="App">
                    <Container>
                       <Header cams={cams} current={this.state.current}/>


                        <Route exact path="/">
                            <CameraSummary cameras={cams} />
                        </Route>
                        <Route path="/cameras/:id">
                            <Camera onShow={this.onShowCamera.bind(this)}/>
                        </Route>
                    </Container>
                </div >
            </Router >
        );
    }
}



function Camera(props) {
    // // We can use the `useParams` hook here to access
    // // the dynamic pieces of the URL.
    let { id } = useParams();

    //props.onShow(id);
    return (
        <CameraView cameraid={id} />
    );
}


export default App;

import React, { Component } from 'react';
import { BrowserRouter, Switch } from 'react-router-dom';
import ReactDOM from 'react-dom';
import Passcode from './passcode';
import Twofactor from './twofactor';
import TerminalTab from './terminaltab';
import {
    BrowserRouter as Router,
    Route,
    Link,
    Redirect,
    withRouter
  } from "react-router-dom";
import './styles.css';
import auth from './auth';
import OTP from './otp';
//import Terminal  from "../../../node_modules/xterm/dist/xterm";

class App extends Component {

    public render() {
        return (
            <Router>
                <Route exact path="/" component={Passcode} />
                <Route path="/two-factor" component={Twofactor}/>
                <Route path="/terminal" component={TerminalTab} />
            </Router>
        )
    }
};

ReactDOM.render(
    <App />,
    document.getElementById('root'),
);
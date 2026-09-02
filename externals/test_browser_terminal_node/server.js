"use strict";
exports.__esModule = true;
var express = require("express");
var pty = require("node-pty");
var app = express();
var expressWs = require('express-ws')(app);
// Serve static assets from ./static
app.use(express.static(__dirname + "/static"));
// Instantiate shell and set up data handlers
expressWs.app.ws('/ws', function (ws, req) {
    // Spawn the shell
    var shell = pty.spawn('/bin/bash', [], {
        name: 'xterm-color',
        cwd: process.env.PWD,
        env: process.env
    });
    // For all shell data send it to the websocket
    shell.on('data', function (data) {
        console.log(data);
        ws.send(data);
    });
    // For all websocket data send it to the shell
    ws.on('message', function (msg) {
        console.log(msg);
        shell.write(msg);
    });
});
// Start the application
app.listen(12345);

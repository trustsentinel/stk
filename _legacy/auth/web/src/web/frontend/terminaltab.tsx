import React, { useState , useEffect, useRef}  from 'react';
import remote from './terminal';
import './terminaltab.css';
import { confirmAlert } from 'react-confirm-alert'; // Import
import 'react-confirm-alert/src/react-confirm-alert.css'; // Import css
import globals from './globals'
import utils from './utils';

let fetched = false

const channelAddress = (token) => "ws://localhost:8081/channels/" + token + "/ws"
const registered =  globals.backend + "/devices/registered"

export class Channel {
    
    socket = null
    onmessage = (channel :Channel, message: any) => {}
    
    onMessage = (evt: any) => {
        console.log("onmessage", evt)
        if (this.onmessage){
            this.onmessage(this, evt.data)
        }
    }

    open(options : any, callbackOnOpen: () => void, callbackOnMessage: (channel :Channel, message: any) => void){
        this.socket = new WebSocket(options.address)
        this.socket.onmessage = this.onMessage
        this.socket.onopen = callbackOnOpen
        this.onmessage= callbackOnMessage
    }

    write(type, message: any){
        this.socket.send(JSON.stringify({
            type: type,
            data: message
        }))
    }

    close(){
        this.socket.close()
    }
} 


const TerminalTab = (_) => {

    let container;

    const isEncryptedEnabled = globals.encrypted
    console.log("!!!!!!!!!!!!!!!!!", globals)

    const [message, setMessage] = useState("Please select a terminal to connect!")
    const [devices, setDevices] = useState([ { token: "device1" }, { token: "device2" }]);

    if (!fetched){
        utils.get(registered, 
            (data) => {
                setDevices(data)
                fetched = true
            });
    }

    const authentication = (token, callback) => {
        let sid = 0
        let auth = new Channel()
        auth.open({
            address: globals.backend.replace("http", "ws") + "/devices/" + token
        }, () => {
            if (callback) callback()
        }, (ch, data) => {
            const message = JSON.parse(data)
            switch (message.Type){
                case "auth":
                    if (message.Data["sid"]){
                        sid = message.Data["sid"]
                        alert("Sid:" + sid)
                        ch.write("auth", {
                            key: "00000",
                            authdata: "allow_me_2_enter",
                            sid: sid
                        })
                    }
                    break;
            }
        })
    }

    const sleep = (milliseconds) => {
        return new Promise(resolve => setTimeout(resolve, milliseconds))
      }

    const openAndConnectServer  = (_: any, token: string) => {
        setMessage("Authentication...")
        authentication(token, () => {
            setMessage("Authentication done.....")
        })
        sleep(20000)
        setMessage("Connecting to terminal " + token + " ...")
        sleep(1000)
        remote.init(container);
        remote.openAndConnect({address: channelAddress(token)}, () =>  {
            sleep(3000)
            setMessage("Connected.")
        })
        remote.onUpdateStatus(setMessage)
    }

    const onSelectedTerminal = (index: number, token: string) => () => {
        utils.confirmation('Connecting to terminal ' + token, 
            () => openAndConnectServer(index, token), 
            () => {})
    }

    return (
        <div className="terminal-selector page">
           
            { isEncryptedEnabled ? null : <span style={{float: "right",color: "white", fontSize: "12px", position: "relative", top: "-92px", right: "-218px"}}>Encryption disabled</span>}
           
            { devices.map((_, index) => <input type="radio" key={"input-" + index} name="tab" id={"tab-" + index} ></input>)}
            
            <div className="tabs-wrapper">
            { devices.map((device, index) => <label htmlFor={"tab-" + index} 
                key={"label-" + index} className={"label-tab-" + index}
                onClick={onSelectedTerminal(index, device.token)}>{device.token}</label>)}
            </div>  
            
            <div id="terminal" className="terminal-container">
                <span>{message}</span>
                <div ref={ el => container = el } className="terminal"></div>
            </div>
        </div>
    );
};

export default TerminalTab;
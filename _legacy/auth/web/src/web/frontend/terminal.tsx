import { FitAddon } from 'xterm-addon-fit';
import { Terminal as TerminalType, Terminal} from 'xterm'
import globals from './globals'
import {A, B, SecureChannel, Writer} from './secure'


class TerminalWriter implements Writer {
    terminal = null
    constructor(terminal){
        this.terminal = terminal
    }
    write(data: any){
        this.terminal.write(data)
    }
}

class RemoteTerminal {

    connected = false
    terminal = null
    fitAddon = null
    container = null
    status = ""

    channel: SecureChannel

    // listeners
    onUpdateStatusListener = (any) => {}

    constructor() {
        this.terminal = new Terminal({
            cursorBlink: true,
            rows: 20,
            cols: 120,
            fontSize: 14
          })
        this.setStatus("initialised")
        const typedTerm = this.terminal as TerminalType;
        this.fitAddon = new FitAddon();
        typedTerm.loadAddon(this.fitAddon);
    }

    init(container){
        this.container = container
    }

    onUpdateStatus(cb){
        this.onUpdateStatusListener = cb
    }

    setStatus(status: string){
        this.onUpdateStatus(status)
        this.status = status;
    }

    write(msg: any) {
        console.log("[socket] writing: " , msg)
        this.terminal.write(msg);
    }

    isConnected(){
        return this.connected;
    }

    initialisation() {
        
        const term = this.terminal;
        term.open(this.container)
        term.element.style.padding = '15px';
        this.fitAddon.fit();
        term.focus();


        const writer = new TerminalWriter(this.terminal)
        const channel = new SecureChannel(writer)

        term.prompt = () => {
            term.write('\r\n$ ');
        }
        term.prompt()
        term.onKey((e: { key: string, domEvent: KeyboardEvent }) => {
            const ev = e.domEvent;
            const printable = !ev.altKey && !ev.ctrlKey && !ev.metaKey;
        
            if (ev.keyCode === 13) {
                term.prompt();
                channel.write('\r\n')
            } else if (ev.keyCode === 8) {
                if (term._core.buffer.x > 2) {
                    channel.write('\b \b')
                }
            } else if (printable) {
                channel.write(e.key)
            }
        });
        this.channel = channel;
    }

    connect(address: string, cb: () => void){
        this.setStatus("Connecting to " + address)
        console.log("***********************")
        console.log("REMOTE KEY", B.publicKey)
        console.log("LOCAL KEY", A.publicKey)
        console.log("***********************")
        const options = {
            address: address,
            encryption: globals.encrypted,
            K: B.publicKey
        }
        this.channel.open(options, () => {
            console.log("Connection established")
        })
        this.setStatus("Connected to " + address)
        cb()
    }

    getStatus(){
        return this.status
    }

    openAndConnect(options: any, cb: () => void) {
        setTimeout(() => {
            this.initialisation()
            const address = (!options.address)  ? globals.backend.replace("http", "ws") + "/channels/dcb6888d/ws": options.address;
            this.setStatus("Connecting to " + address)
            this.connect(address, () => {
                this.connected = true;
                cb();
            })
        },1500)
    }

}

export default new RemoteTerminal();
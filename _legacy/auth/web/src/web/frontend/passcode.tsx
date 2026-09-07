import React from 'react';
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom'
import auth from './auth';
import { isNumber } from 'util';



const Passcode = (props: { history: string[]}) => {
    console.log(props)

    const onChangePasscode = (evt) => {
        const code = evt.target.value.replace(' ','')
        if (!isNaN(Number(code)) && code.length <= 6){
            passcode = code.replace(/^([\d]{3}?)/,'$1 ')
        }
        evt.target.value = passcode
    }
    let passcode = ""

    const handleLogin = (value) => {
        const passcode = value.replace(' ','')
        auth.login(passcode, ()=> {
            if (passcode == "111111"){
                console.log("Authenticated!")
                props.history.push('/two-factor');
                return;
            }
            console.log("Not Authenticated!")
        })
    }
    
    return (
        <div className="login page">
            <div className="form">
                <form className="passcode-form" onSubmit={() => handleLogin(passcode)}>
                    <input type="text" placeholder="" 
                        onChange={onChangePasscode}/>
                    <button type="submit" id="submit-passcode">Submit passcode</button>
                </form>
            </div>
        </div>
    )
};

export default Passcode;
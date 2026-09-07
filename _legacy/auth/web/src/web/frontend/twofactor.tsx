

import React, { useState }  from 'react';
import auth from './auth';
import globals from './globals'

const Twofactor = (props: { history: string[]; }) => {    
    let totp = ""

    const [image, setImage] = useState(globals.frontend + "/assets/generation.png");
    
    const generateQrUrl = (passcode) => {
        return globals.frontend + "/auth/qr/" + passcode
    }

    const handleGenerateQr = () => {
        if (auth.passcode && auth.passcode.length == 6){
            let img = generateQrUrl(auth.passcode);
            setImage(img)
        }
    }

    const handleTotp = () => {
        const totpCode = totp.replace(' ','')
        auth.totpValidation(totpCode, ()=> {
            console.log("Totp validated!")
        })
        props.history.push('/terminal');
    }

    const onChangePassword = (evt) => {
        const code = evt.target.value.replace(' ','')
        if (!isNaN(Number(code)) && code.length <= 6){
            totp = code.replace(/^([\d]{3}?)/,'$1 ')
        }
        evt.target.value = totp
    }
  
    return (
        <div className="login page">
            <div className="form">
                <span style={{ fontSize:"12px"}}>Please connect Yubikey to perform U2F syncronisation otherwise</span>
                <form className="otp-form">
                    <img onClick={handleGenerateQr} src={image} width="100%" style={{ padding:"23px", width:"190px"}}/>
                    <input type="text" placeholder="otp" onChange={onChangePassword}/>
                    <button type="button" id="otp-code" onClick={() => handleTotp()}>Enter OTP code</button>
                </form>
            </div>
        </div>
    )
};

export default Twofactor;
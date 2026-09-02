const TOKEN_KEY = 'jwt2';

import globals from './globals'

class Auth {

    authenticated = false
    loggedIn = false
    passcode = null

    constructor() {
        this.loggedIn = false;
        this.authenticated = false;
        this.passcode = null;
    }

    login(passcode, cb: () => void) {
        localStorage.setItem(TOKEN_KEY, passcode);
        this.loggedIn = true;
        this.passcode = ""+passcode;
        cb();
    }

    totpRegistration(cb: () => void) {
        localStorage.setItem(TOKEN_KEY, this.passcode);
        cb();
    }

    totpValidation(totp: any, cb: (response: any) => void) {
        fetch(globals.backend + "/auth/validate/"+this.passcode+"/" + totp)
        .then(response => response.json())
        .then(_ => cb);
        this.authenticated = true;
    }

    logout(cb: () => void) {
        localStorage.removeItem(TOKEN_KEY);
        this.loggedIn = false;
        this.authenticated = false;
        cb();
    }

    isAuthenticated() {
        return this.authenticated;
    }

    isLoggedIn() {
        return this.loggedIn;
    }

}

export default new Auth();
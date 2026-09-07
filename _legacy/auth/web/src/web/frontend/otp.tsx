class OTP {

    authenticated = false;

    constructor() {
        this.authenticated = false;
    }

    login(cb: () => void) {
        this.authenticated = true;
        cb();
    }

    logout(cb: () => void) {
        this.authenticated = false;
        cb();
    }

    isAuthenticated() {
        console.log(this);
        return this.authenticated;
    }

}

export default new OTP();
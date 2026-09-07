

export default {
    backend: process.env.BACKEND ? "http://" + process.env.BACKEND : "http://localhost:8081",
    frontend: process.env.FRONTEND ? "http://" + process.env.FRONTEND : "http://localhost:9999",
    encrypted: process.env.ENCRYPTED ? true : false
}

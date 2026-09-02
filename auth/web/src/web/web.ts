import express from "express";
import http from "http";
import path from "path";

// Express app initialization
const app = express();

console.log("*******************************")
console.log("Encrypted:  " + process.env.ENCRYPTED ? "enabled": "disabled")
console.log("Backend:    " + process.env.BACKEND ? process.env.BACKEND : "localhost:8081")
console.log("Frontend:   " + process.env.FRONTEND ? process.env.FRONTENDFr : "localhost:9999")
console.log("*******************************")

// Template configuration
app.set("view engine", "ejs");
app.set("views", "public");

// Static files configuration
app.use("/assets", express.static(path.join(__dirname, "frontend")));

// Controllers
app.get("/*", (req, res) => {
    res.render("index");
});

// Start function
export const start = (port: number): Promise<void> => {
    const server = http.createServer(app);

    return new Promise<void>((resolve, reject) => {
        server.listen(port, resolve);
    });
};
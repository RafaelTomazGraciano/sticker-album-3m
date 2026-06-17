import {WebSocketServer} from "ws";
import {handleMiddle} from "./middle";

const wss = new WebSocketServer({port: 8080});

wss.on("connection", (ws) => {

    ws.on("message", handleMiddle);

});
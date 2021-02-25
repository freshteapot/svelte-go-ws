import { writable } from 'svelte/store';

// https://gist.github.com/jfromaniello/8418116
// https://svelte.dev/repl/29a5bdfb981f479fb387298aef1190a0?version=3.22.2

const messageStore = writable('');

// TODO use wss for production (get from meta data)
const socket = new WebSocket('ws://localhost:8080/echo');

function toEvent(message) {
    try {
        var event = JSON.parse(message.data);
        const appEvent = new CustomEvent(event.type, { detail: event.data });
        socket.dispatchEvent(appEvent);

    } catch (err) {
        console.log('not an event', err);
    }
}

// Connection opened
socket.addEventListener('open', function (event) {
    console.log("It's open");
});

socket.addEventListener('close', function (event) {
    console.log("It's closed");
});

socket.addEventListener('authenticate', function (event) {
    console.log("authenticate", event);
    // TODO get from localstorage
    socket.send(JSON.stringify({
        type: "authenticate",
        data: JSON.stringify({
            "bearer": "XXX",
            "user_uuid": "XXX"
        })
    }));
});

socket.addEventListener('authenticated', function (event) {
    console.log("authenticated", event);
});


// Listen for messages
socket.addEventListener('message', toEvent)

socket.addEventListener('update', (event) => {
    alert(event.detail)
});

socket.addEventListener('chat', (event) => {
    messageStore.set(event.detail);
});

const sendMessage = (message) => {
    if (socket.readyState <= 1) {
        socket.send(JSON.stringify({
            type: "message",
            data: message,
        }));
    }
}


export default {
    subscribe: messageStore.subscribe,
    sendMessage
}

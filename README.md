# Svelte + websocket + golang + authentication

# What
- Svelte frontend
- Vanilla websocket
- Golang server to provide echo with an attempt at security in the first message

## Get started
- 2 terminals
- 1 browser
- Learnalist credentials optional for the authentication step

## Frontend
Install the dependencies...

```bash
npm install
```

...then start [Rollup](https://rollupjs.org):

```bash
npm run dev
```

## Backend
```
go run main.go
```

# Browser
Goto "http://localhost:8080/"


# Reference

- https://blog.alexellis.io/golang-json-api-client/
- https://gist.github.com/jfromaniello/8418116
- https://svelte.dev/repl/29a5bdfb981f479fb387298aef1190a0?version=3.22.2
- https://developer.mozilla.org/en-US/docs/Web/Guide/Events/Creating_and_triggering_events
- https://devcenter.heroku.com/articles/websocket-security
- https://svelte.dev
- https://github.com/freshteapot/learnalist-api/blob/master/docs/api.user.info.md

# markdown editor

Markdown editor to use with Websocket server [here](https://github.com/iamkahvi/go_markdown_server).

## Development
`npm run start` serves the website on http://localhost:3000.

## api design

```ts
type Patch = [-1 | 0 | 1, string];

interface Message {
  patches: Patch[];
}

interface MyResponse {
  status: "OK" | "ERROR";
  doc?: string;
}

// ex of the first message
const m: Message = {
    patches: []
}

// ex of the first response
const r: MyResponse = {
    status: "OK",
    doc: "hello world"
}

// ex of normal message
const m2: Message = {
    patches: [
        [0, "hello world"],
        [1,"!"]
    ]
}

// ex of normal response
const r2: MyResponse = {
    status: "OK"
}
```


## server side logic

```ts
const fullString = patches.reduce((acc: string, [op, text]: Patch) => {
  if (op === 1) {
    return acc + text;
  } else if (op === 0) {
    return acc + text;
  } else {
    return acc;
  }
}, "");
```

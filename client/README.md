# markdown editor

Markdown editor to use with the websocket server in `/server`.

## development
`npm run start` serves the website on http://localhost:3000

## api design
```ts
type Diff = [number, string];

type PatchObj = {
  diffs: Diff[];
  start1: number | null;
  start2: number | null;
  length1: number;
  length2: number;
};

interface Message {
  patches: PatchObj[];
}

interface MyResponse {
  status: "OK" | "ERROR";
  doc?: string;
}
```

### example flow

```ts
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
    {
      "diffs": [
        [0, "lo world"],
        [1, "!"]
      ],
      "start1": 3,
      "start2": 3,
      "length1": 8,
      "length2": 9
    }
  ]
};
```

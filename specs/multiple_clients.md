# multiple clients

## v1

- [x] `clients: <num>` updating counter on the client

## v2

- [x] ability to "broadcast" event across all channels (client # decreasing)

## v3: dynamic editor

1. client a connects as EDITOR
2. client b connects as READER
 - client b's editor is in read-only state
4. client a diconnects
5. client b becomes EDITOR

## handler approach

- spin up a dedicated goroutine that continuously calls `c.ReadJSON(&m)` and forwards each payload into a buffered `connCh`
- maintain the main handler loop with a `select` that listens to `connCh`, the broker subscription channel, an error channel
- ensure only the main handler goroutine performs websocket writes; forward outbound broker messages through the same loop so writes remain serialized
- cancel the context and unsubscribe to stop the reader goroutine cleanly when the handler exits

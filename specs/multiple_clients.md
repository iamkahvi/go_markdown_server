# multiple clients

## v1

- [x] `clients: <num>` updating counter on the client

## v2

- [ ] ability to "broadcast" event across all channels (client # decreasing)

## handler approach

- spin up a dedicated goroutine that continuously calls `c.ReadJSON(&m)` and forwards each payload into a buffered `connCh`
- maintain the main handler loop with a `select` that listens to `connCh`, the broker subscription channel, an error channel
- ensure only the main handler goroutine performs websocket writes; forward outbound broker messages through the same loop so writes remain serialized
- cancel the context and unsubscribe to stop the reader goroutine cleanly when the handler exits

```go
ctx, cancel := context.WithCancel(req.Context())
defer cancel()

connCh := make(chan []byte, 1)   // drain ws reads
errCh := make(chan error, 1)     // propagate read errors

go func() {
    defer close(connCh)
    for {
        _, data, err := conn.ReadJSON()
        if err != nil {
            errCh <- err
            return
        }
        select {
        case connCh <- data:      // deliver to handler loop
        case <-ctx.Done():
            return
        }
    }
}()

subCh := broker.Subscribe()
defer broker.Unsubscribe(subCh)

for {
    select {
    case payload := <-connCh:
        // handle inbound websocket message
    case msg := <-subCh:
        // handle broker message (fan-out)
    case err := <-errCh:
        // log and break on websocket read failure
        return err
    case <-ctx.Done():
        return
    }
}
```

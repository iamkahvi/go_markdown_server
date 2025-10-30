# go_markdown_server
Websocket server to use with markdown editor in browser [here](https://github.com/iamkahvi/markdown_editor).

## Development
`go run main.go` listens on port 8000

`npx wscat -c ws://localhost:8000/write -H Origin:https://write.kahvipatel.com` then paste
`{"type":"normal","status":"success","data":"this is kahvi"}`

## Routes
`/` - homepage  
`/write` - websocket

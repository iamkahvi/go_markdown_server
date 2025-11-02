# go_markdown_server
Websocket server to use with markdown editor in browser [here](https://github.com/iamkahvi/markdown-editor).

## development
- if you have nix installed `nix develop` listens on port 8000
- if not, try `DEV=1 go run main.go`

## routes
`/` - homepage
`/write` - websocket

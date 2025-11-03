# go_markdown_server
Websocket server to use with the markdown editor in `/client`.

## development
- if you have nix installed `nix develop` listens on port 8000
- run the development server with `DEV=1 go run ./cmd/server`
- alternatively, use `make run` to execute the same command

## routes
`/` - homepage
`/write` - websocket

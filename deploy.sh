#!/bin/bash

env GOOS=linux GOARCH=386 go build -v . && scp text_editor_server root@142.93.152.73:/root/text_editor_server

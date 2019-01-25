#!/bin/bash
GOOS=windows GOARCH=386 go build  -o bitballot_win.exe -ldflags "-s -w" && upx bitballot_win.exe && mv bitballot_win.exe app/.

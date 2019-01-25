#!/bin/bash
#upx bitballot_linux.elf &&
GOOS=linux GOARCH=amd64 go build -o bitballot_linux.elf -ldflags "-s -w" && upx bitballot_linux.elf && mv bitballot_linux.elf app/.

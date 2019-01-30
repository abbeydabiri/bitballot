#!/bin/bash
#upx bitballot_linux.elf &&
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bitballot_linux.elf -ldflags "-s -w" && upx bitballot_linux.elf

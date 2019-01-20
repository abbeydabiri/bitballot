#!/bin/bash
GOOS=linux GOARCH=386 go build  -o bitballot.elf -ldflags "-s -w" && upx bitballot.elf
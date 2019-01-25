#!/bin/bash
#go build  -o bitballot_mac.app -ldflags "-s -w" && mv bitballot_mac.app app/.
#go build  -o bitballot_mac.app -ldflags "-s -w" && upx "-9" bitballot_mac.app && mv bitballot_mac.app app/.
go build  -o bitballot_mac.app -ldflags "-s -w"

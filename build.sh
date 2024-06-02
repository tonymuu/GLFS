#!/bin/bash
go clean

# build the filesystem
go build -C ./filesystem/ -o ../build/glfs

# build the application
go build -C ./application/ -o ../build/app
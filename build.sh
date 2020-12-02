#! /bin/bash

go build -ldflags "-s -w"
npm run build

#! /bin/bash

npm i
go build -ldflags "-s -w"
npm run build

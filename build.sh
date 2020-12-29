#! /bin/bash

npm i --no-package-lock
go build -ldflags "-s -w"
npm run build

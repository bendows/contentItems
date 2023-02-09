#!/bin/sh
go build || exit 1
git add .
git commit -m ok
git push origin master

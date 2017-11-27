#!/usr/bin/env bash

git clone https://github.com/lazerion/hz-go-it
ls -al
cd ./src/github.com/lazerion/acceptance/
go build
go test
rc=$?
if [[ ${rc} -ne 0 ]] ; then
  echo 'could not perform tests'; exit $rc
fi
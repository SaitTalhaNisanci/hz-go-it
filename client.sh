#!/usr/bin/env bash

git clone https://github.com/lazerion/hz-go-it
ls -al
cd ./acceptance/
go get -u all
go build
go test
rc=$?
if [[ ${rc} -ne 0 ]] ; then
  echo 'could not perform tests with success'; exit $rc
fi
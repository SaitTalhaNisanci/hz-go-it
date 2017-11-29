#!/usr/bin/env bash

git clone https://github.com/lazerion/hz-go-it
cd ./hz-go-it/acceptance/
ls -al
go env
go test -run TestSingleMemberConnection
rc=$?
if [[ ${rc} -ne 0 ]] ; then
  echo 'could not perform tests with success'; exit $rc
fi
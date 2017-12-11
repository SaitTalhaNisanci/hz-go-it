#!/usr/bin/env bash

git clone https://github.com/lazerion/hz-go-it
# uncomment below to test changes on local rather than pushing
#cd /local/source
cd ./hz-go-it/acceptance/
go test -run TestSingleMemberConnection
rc=$?
if [[ ${rc} -ne 0 ]] ; then
  echo 'could not perform tests with success'; exit $rc
fi
#!/bin/sh

rm logs/*
make test-build
ENABLE_DEBUG=${ENABLE_DEBUG:-0}
DYLD_LIBRARY_PATH=${ACLDIR}/lib
CURDIR=`pwd`
TESTFILES=${CURDIR}/files
LOGPATH=${CURDIR}/logs
REFERENCES=${TESTFILES}/references

export DYLD_LIBRARY_PATH ENABLE_DEBUG CURDIR
export TESTFILES LOGPATH REFERENCES

GOOS=`go env GOOS`
GOARCH=`go env GOARCH`
TEST_EXECUTE=bin/tests/${GOOS}_${GOARCH}/adabas.test

if [ $# -gt 1 ]; then
   TEST_EXECUTE=bin/tests/${GOOS}_${GOARCH}/$1.test
   shift
fi
PARA=
if [ $# -gt 0 ]; then
   if [ ! $1 = "all" ]; then
      PARA="-test.run $1"
   fi
else
   echo "Start all tests"
fi

${TEST_EXECUTE} -test.v ${PARA}

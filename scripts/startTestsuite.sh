#!/bin/sh

ulimit -c unlimited

ADABAS_ACCESS_HOME=`pwd`

DYLD_LIBRARY_PATH=:/Volumes/SAG-Q/testenv/adav67d/AdabasClient/lib:/lib:/usr/lib
export DYLD_LIBRARY_PATH
ENABLE_DEBUG=${ENABLE_DEBUG:-0}
LOGPATH=`pwd`/logs
TESTFILES=`pwd`/files
GO_ADA_MESSAGES=`pwd`/messages
GOPATH=/tmp/tmp_gopath.`id -u`:$GOPATH
export ENABLE_DEBUG LOGPATH TESTFILES GO_ADA_MESSAGES GOPATH
rm -f ./logs/*.log

if [ ! "$ACLDIR" == "" ]; then
  GO_TAGS="release adalnk"
  CGO_CFLAGS="-DCE_T${SAGTARGET} -I${ADABAS_ACCESS_HOME}/c/SAGENV -I${ADABAS_ACCESS_HOME}/c -I${ACLDIR}/inc"
  CGO_LDFLAGS="-L${ACLDIR}/lib -ladalnkx" 
  export CGO_CFLAGS CGO_LDFLAGS
  go run -tags "$GO_TAGS" ${TESTS_RUN} -v tests/testsuite/main.go $*
else
  GO_TAGS="release"
  go run -tags "$GO_TAGS" ${TESTS_RUN} -v tests/testsuite/main.go $*
fi


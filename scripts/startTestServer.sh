#!/bin/sh

if [ $# -eq 1 ]; then
   echo "Start test $1"
   TESTS_RUN="-run $1"
else
   TESTS_RUN=
fi
ADABAS_ACCESS_HOME=`pwd`

DYLD_LIBRARY_PATH=:/Volumes/SAG-Q/testenv/adav67/AdabasClient/lib:/lib:/usr/lib
export DYLD_LIBRARY_PATH
ENABLE_DEBUG=1
LOGPATH=`pwd`/logs
export ENABLE_DEBUG LOGPATH
rm -f ./logs/*.log
CGO_CFLAGS="-DCE_T${SAGTARGET} -I${ADABAS_ACCESS_HOME}/c/SAGENV -I${ADABAS_ACCESS_HOME}/c" CGO_LDFLAGS="-L${ACLDIR}/lib -ladalnkx" GOPATH=`pwd`:$GOPATH go test ${TESTS_RUN} -v adabas/server


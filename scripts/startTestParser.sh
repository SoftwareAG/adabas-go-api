#!/bin/sh

if [ $# -eq 1 ]; then
   echo "Start test $1"
   TESTS_RUN="-run $1"
else
   TESTS_RUN=
fi
ADABAS_ACCESS_HOME=`pwd`

DYLD_LIBRARY_PATH=:/Volumes/SAG-Q/testenv/adav67d/AdabasClient/lib:/lib:/usr/lib
export DYLD_LIBRARY_PATH
rm -f ./logs/*.log
ENABLE_DEBUG=${ENABLE_DEBUG:0}
LOGPATH=`pwd`/logs
export ENABLE_DEBUG LOGPATH
CGO_CFLAGS="-DCE_T${SAGTARGET} -I${ADABAS_ACCESS_HOME}/c/SAGENV -I${ADABAS_ACCESS_HOME}/c" CGO_LDFLAGS="-L${ACLDIR}/lib -ladalnkx" GOPATH=/tmp/tmp_adabas-go-api.$(id -u):$GOPATH go test ${TESTS_RUN} -v github.com/SoftwareAG/adabas-go-api/adatypes


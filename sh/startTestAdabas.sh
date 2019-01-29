#!/bin/sh

ulimit -c unlimited

if [ $# -eq 1 ]; then
   echo "Start test $1"
   TESTS_RUN="-run $1"
else
   TESTS_RUN=
fi
ADABAS_ACCESS_HOME=`pwd`

DYLD_LIBRARY_PATH=:/Volumes/SAG-Q/testenv/adav67/AdabasClient/lib:/lib:/usr/lib
export DYLD_LIBRARY_PATH
ENABLE_DEBUG=${ENABLE_DEBUG:-0}
ADAMFDBID=54712
ADATCPHOST=${ADATCPHOST:-emon:60177}
LOGPATH=`pwd`/logs
TESTFILES=`pwd`/files
REFERENCES=${TESTFILES}/references
REFERENCE_WRITE=
GO_ADA_MESSAGES=`pwd`/messages
export ENABLE_DEBUG LOGPATH TESTFILES GO_ADA_MESSAGES REFERENCES REFERENCE_WRITE ADAMFDBID
export ADATCPHOST
rm -f ./logs/*.log
CGO_CFLAGS="-DCE_T${SAGTARGET} -I${ADABAS_ACCESS_HOME}/c/SAGENV -I${ADABAS_ACCESS_HOME}/c -I${ACLDIR}/inc" CGO_LDFLAGS="-L${ACLDIR}/lib -ladalnkx" GOPATH=/tmp/tmp_adabas-go-api.$(id -u):$GOPATH go test ${TESTS_RUN} -count=1 -v -tags "release adalnk" github.com/SoftwareAG/adabas-go-api/adabas


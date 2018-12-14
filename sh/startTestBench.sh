#!/bin/sh

if [ $# -eq 1 ]; then
   echo "Start test $1"
   TESTS_RUN="-run XXX -bench $1"
else
   TESTS_RUN=
fi
ADABAS_ACCESS_HOME=`pwd`

DYLD_LIBRARY_PATH=:/Volumes/SAG-Q/testenv/adav67/AdabasClient/lib:/lib:/usr/lib
export DYLD_LIBRARY_PATH
rm -f ./logs/*.log
CGO_CFLAGS="-I${ACLDIR}/inc" CGO_LDFLAGS="-L${ACLDIR}/lib -ladalnkx" GOPATH=/tmp/tmp_gopath.$(id -u):$GOPATH go test ${TESTS_RUN} -v -tags "release adalnk" github.com/SoftwareAG/adabas-go-api/adabas


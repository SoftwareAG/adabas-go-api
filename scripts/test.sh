#!/bin/sh

LD_LIBRARY_PATH="$LD_LIBRARY_PATH:${ACLDIR}/lib"
DYLD_LIBRARY_PATH="$DYLD_LIBRARY_PATH:${ACLDIR}/lib:/lib:/usr/lib"
export LD_LIBRARY_PATH DYLD_LIBRARY_PATH

CGO_CFLAGS="-I${ACLDIR}/inc" 
CGO_LDFLAGS="-L${ACLDIR}/lib -ladalnkx -lsagsmp2 -lsagxts3 -ladazbuf"
export CGO_CFLAGS CGO_LDFLAGS

${GO} test --json -timeout ${TIMEOUT}s -count=1 ${GO_FLAGS} -v $* ./adabas
${GO} test --json -timeout ${TIMEOUT}s -count=1 ${GO_FLAGS} -v $* ./adatypes

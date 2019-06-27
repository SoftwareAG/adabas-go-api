@echo off

set CGO_CFLAGS=-I%ACLDIR%\..\inc 
set CGO_LDFLAGS=-L%ACLDIR% -L%ACLDIR%\..\lib -ladalnkx  
go get golang.org/x/text/encoding
go get github.com\sirupsen\logrus

go build -tags adalnk  -o %GOPATH%/bin/tests/testsuite tests/testsuite.go

go get github.com/stretchr/testify/assert
go get github.com/tebeka/go2xunit


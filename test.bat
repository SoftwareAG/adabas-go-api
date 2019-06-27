@echo off

set CGO_CFLAGS=-I%ACLDIR%\..\inc 
set CGO_LDFLAGS=-L%ACLDIR% -L%ACLDIR%\..\lib -ladalnkx  

mkdir test
go test -timeout 100s -count 1 -tags adalnk -v  github.com/SoftwareAG/adabas-go-api/adabas github.com/SoftwareAG/adabas-go-api/adatypes >test.output

%GOPATH%\bin\go2xunit -input test.output -output test\tests.xml

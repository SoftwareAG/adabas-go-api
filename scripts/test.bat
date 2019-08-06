@echo off

set DIR=%~dp0\..

set CGO_CFLAGS=-I%ACLDIR%\..\inc 
set CGO_LDFLAGS=-L%ACLDIR% -L%ACLDIR%\..\lib -ladalnkx  

set TESTFILES=%DIR%\files
set REFERENCES=%TESTFILES%\references
set LOGPATH=%DIR%\logs

echo "Work in %DIR"
cd %DIR%

if not exist test mkdir test
if not exist %LOGPATH% mkdir %LOGPATH%


mkdir test
go test -timeout 100s -count 1 -tags adalnk -v  ./... >test.output
rem github.com/SoftwareAG/adabas-go-api/adabas github.com/SoftwareAG/adabas-go-api/adatypes

%GOPATH%\bin\go2xunit -input test.output -output test\tests.xml

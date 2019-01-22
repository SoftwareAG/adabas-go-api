@echo off

rem set GOPATH=%cd%\tmp_gopath
mkdir %GOPATH%

set CGO_CFLAGS=-I%ACLDIR%\..\inc 
set CGO_LDFLAGS=-L%ACLDIR% -L%ACLDIR%\..\lib -ladalnkx  
go get golang.org/x/text/encoding
go get github.com\sirupsen\logrus

mkdir %GOPATH%\src\github.com\SoftwareAG\adabas-go-api\ 
xcopy /e /v /I /Y adabas %GOPATH%\src\github.com\SoftwareAG\adabas-go-api\adabas
xcopy /e /v /I /Y adatypes %GOPATH%\src\github.com\SoftwareAG\adabas-go-api\adatypes

go build -tags adalnk  -o %GOPATH%/bin/tests/testsuite tests/testsuite.go

go get github.com/stretchr/testify/assert
go get github.com/tebeka/go2xunit

mkdir test
go test -timeout 100s -tags adalnk -v  github.com/SoftwareAG/adabas-go-api/adabas github.com/SoftwareAG/adabas-go-api/adatypes >test.output

%GOPATH%\bin\go2xunit -input test.output -output test\tests.xml


set GOPATH=%cd%\tmp_gopath
mkdir %GOPATH%

set CGO_CFLAGS=-I%ACLDIR%\..\inc 
set CGO_LDFLAGS=-L%ACLDIR% -L%ACLDIR%\..\lib -ladalnkx  
go get -u golang.org/x/text/encoding
go get -u github.com\sirupsen\logrus
go build -tags adalnk  -o %GOPATH%/bin/tests/testsuite tests/testsuite.go
go test -tags adalnk  -o %GOPATH%/bin/tests/testsuite tests/testsuite.go

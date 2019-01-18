
set GOPATH=C:\Users\tkn\AppData\Local\Temp\tmp_gopath.1065759


set CGO_CFLAGS=-I%ACLDIR%\..\inc 
set CGO_LDFLAGS=-L%ACLDIR% -L%ACLDIR%\..\lib -ladalnkx  
rem go get -u golang.org/x/text/encoding
go get -u github.com\sirupsen\logrus
go build -tags adalnk  -o C:/Users/tkn/Sources/adabas-go-api/bin/tests/testsuite tests/testsuite.go

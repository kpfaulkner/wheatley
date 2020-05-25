REM build pack and publish!
REM go 1.14 has issues building binaries compatible with azure functions.
REM so switch back to 1.13 for this.

set GOROOT=c:\users\kenfa\packages\go-1.13

cd cmd\azurefunction
c:\users\kenfa\packages\go-1.13\bin\go build .
cd ..\..
copy cmd\azurefunction\azurefunction.exe deploy


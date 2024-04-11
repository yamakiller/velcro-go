@echo off 
set GO_COMPILER_PATH=..\..\..\tools\thrift\thrift.exe

@REM %GO_COMPILER_PATH% -r --gen go .\thrift\network\client_id.thrift
@REM copy /y gen-go\network\client_id.go .\network\client_id.go
@REM %GO_COMPILER_PATH% -r --gen go .\thrift\cluster\prvs\alterrule.thrift
%GO_COMPILER_PATH% -r --gen go .\pubs\login.thrift
%GO_COMPILER_PATH% -r --gen go .\pubs\loginservice.thrift

%GO_COMPILER_PATH% -r --gen go .\prvs\request.thrift

md ..\protocols\pubs
copy /y gen-go\protocols\pubs\login.go ..\protocols\pubs\login.go
copy /y gen-go\protocols\pubs\loginservice.go ..\protocols\pubs\loginservice.go
md ..\protocols\prvs
copy /y gen-go\protocols\prvs\request.go ..\protocols\prvs\request.go
@REM copy /y gen-go\protocols\prvs\bundle.go .\cluster\protocols\prvs\bundle.go
@REM copy /y gen-go\protocols\prvs\closing.go .\cluster\protocols\prvs\closing.go
rd/s/q gen-go
echo complate 
pause 
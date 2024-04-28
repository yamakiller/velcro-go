
@echo off 
set GO_COMPILER_PATH=..\..\..\tools\thrift\thrift.exe

%GO_COMPILER_PATH% -r --gen netstd ..\..\..\cluster\proto\pubs\error.thrift
%GO_COMPILER_PATH% -r --gen netstd ..\..\..\cluster\proto\pubs\ping.thrift
%GO_COMPILER_PATH% -r --gen netstd ..\..\..\cluster\proto\pubs\pubkey.thrift
%GO_COMPILER_PATH% -r --gen netstd ..\..\..\cluster\proto\pubs\request.thrift
%GO_COMPILER_PATH% -r --gen netstd .\pubs\login.thrift
%GO_COMPILER_PATH% -r --gen netstd .\pubs\login.thrift
%GO_COMPILER_PATH% -r --gen netstd .\pubs\login.thrift
%GO_COMPILER_PATH% -r --gen netstd .\pubs\login.thrift

%GO_COMPILER_PATH% -r --gen netstd .\pubs\login.thrift
%GO_COMPILER_PATH% -r --gen netstd .\pubs\login_service.thrift
%GO_COMPILER_PATH% -r --gen netstd .\pubs\battle.thrift
%GO_COMPILER_PATH% -r --gen netstd .\pubs\battle_service.thrift
%GO_COMPILER_PATH% -r --gen netstd .\pubs\report_nat_client.thrift

@REM copy /y gen-go\protocols\pubs\login.go ..\protocols\pubs\login.go
@REM copy /y gen-go\protocols\pubs\loginservice.go ..\protocols\pubs\loginservice.go
@REM copy /y gen-go\protocols\pubs\battle.go ..\protocols\pubs\battle.go
@REM copy /y gen-go\protocols\pubs\battle_service.go ..\protocols\pubs\battle_service.go
@REM copy /y gen-go\protocols\pubs\report_nat_client.go ..\protocols\pubs\report_nat_client.go

echo complate 
pause 
@echo off 
set GO_COMPILER_PATH=..\..\..\tools\thrift\thrift.exe

@REM %GO_COMPILER_PATH% -r --gen go .\thrift\network\client_id.thrift
@REM copy /y gen-go\network\client_id.go .\network\client_id.go
@REM %GO_COMPILER_PATH% -r --gen go .\thrift\cluster\prvs\alterrule.thrift
%GO_COMPILER_PATH% -r --gen go .\pubs\login.thrift
%GO_COMPILER_PATH% -r --gen go .\pubs\login_service.thrift
%GO_COMPILER_PATH% -r --gen go .\pubs\battle.thrift
%GO_COMPILER_PATH% -r --gen go .\pubs\battle_service.thrift
%GO_COMPILER_PATH% -r --gen go .\pubs\report_nat_client.thrift


@REM %GO_COMPILER_PATH% -r --gen go .\prvs\request.thrift
@REM %GO_COMPILER_PATH% -r --gen go .\prvs\exitbattle.thrift
@REM %GO_COMPILER_PATH% -r --gen go .\prvs\report_nat.thrift
@REM rd/s/q ..\protocols
md ..\protocols\pubs
copy /y gen-go\protocols\pubs\login.go ..\protocols\pubs\login.go
copy /y gen-go\protocols\pubs\login_service.go ..\protocols\pubs\login_service.go
copy /y gen-go\protocols\pubs\battle.go ..\protocols\pubs\battle.go
copy /y gen-go\protocols\pubs\battle_service.go ..\protocols\pubs\battle_service.go
copy /y gen-go\protocols\pubs\report_nat_client.go ..\protocols\pubs\report_nat_client.go

md ..\protocols\prvs
@REM copy /y gen-go\protocols\prvs\request.go ..\protocols\prvs\request.go
@REM copy /y gen-go\protocols\prvs\bundle.go .\cluster\protocols\prvs\bundle.go
@REM copy /y gen-go\protocols\prvs\closing.go .\cluster\protocols\prvs\closing.go
@REM copy /y gen-go\protocols\prvs\exitbattle.go ..\protocols\prvs\exitbattle.go
@REM copy /y gen-go\protocols\prvs\report_nat.go ..\protocols\prvs\report_nat.go
rd/s/q gen-go
echo complate 
pause 
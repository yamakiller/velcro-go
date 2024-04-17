@echo off 
set GO_COMPILER_PATH=tools\thrift\thrift.exe

@REM rpc
@REM %GO_COMPILER_PATH% -r --gen go .\rpc\messages\proto\rpc_ping.thrift
@REM copy /y gen-go\rpc\messages\rpc_ping.go .\rpc\messages\rpc_ping.go

@REM %GO_COMPILER_PATH% -r --gen go .\rpc\messages\proto\rpc_request.thrift
@REM copy /y gen-go\rpc\messages\rpc_request.go .\rpc\messages\rpc_request.go

%GO_COMPILER_PATH% -r --gen go .\rpc\messages\proto\rpc_response.thrift
copy /y gen-go\rpc\messages\rpc_response.go .\rpc\messages\rpc_response.go

@REM %GO_COMPILER_PATH% -r --gen go .\rpc\messages\proto\rpc_service.thrift
@REM copy /y gen-go\rpc\messages\rpc_service.go .\rpc\messages\rpc_service.go

@REM network
@REM %GO_COMPILER_PATH% -r --gen go .\network\proto\client_id.thrift
@REM copy /y gen-go\network\client_id.go .\network\client_id.go

@REM cluster
@REM %GO_COMPILER_PATH% -r --gen go .\cluster\proto\prvs\alterrule.thrift
@REM %GO_COMPILER_PATH% -r --gen go .\cluster\proto\prvs\bundle.thrift
@REM %GO_COMPILER_PATH% -r --gen go .\cluster\proto\prvs\closing.thrift
@REM %GO_COMPILER_PATH% -r --gen go .\cluster\proto\prvs\service.thrift

@REM %GO_COMPILER_PATH% -r --gen go .\cluster\proto\pubs\error.thrift
@REM %GO_COMPILER_PATH% -r --gen go .\cluster\proto\pubs\ping.thrift
@REM %GO_COMPILER_PATH% -r --gen go .\cluster\proto\pubs\pubkey.thrift
@REM %GO_COMPILER_PATH% -r --gen go .\cluster\proto\pubs\request.thrift

@REM copy /y gen-go\protocols\prvs\alterrule.go .\cluster\protocols\prvs\alterrule.go
@REM copy /y gen-go\protocols\prvs\bundle.go .\cluster\protocols\prvs\bundle.go
@REM copy /y gen-go\protocols\prvs\closing.go .\cluster\protocols\prvs\closing.go
@REM copy /y gen-go\protocols\prvs\service.go .\cluster\protocols\prvs\service.go

@REM copy /y gen-go\protocols\pubs\error.go .\cluster\protocols\pubs\error.go
@REM copy /y gen-go\protocols\pubs\ping.go .\cluster\protocols\pubs\ping.go
@REM copy /y gen-go\protocols\pubs\pubkey.go .\cluster\protocols\pubs\pubkey.go
@REM copy /y gen-go\protocols\pubs\request.go .\cluster\protocols\pubs\request.go
@REM %GO_COMPILER_PATH% -r --gen go .\thrift\rpc\messages\rpc_service.thrift


rd/s/q gen-go
echo complate 
pause 
@echo off 
set GO_COMPILER_PATH=tools\thrift\thrift.exe

@REM rpc
%GO_COMPILER_PATH% -r --gen go .\rpc\messages\proto\rpc_ping.thrift
copy /y gen-go\rpc\messages\rpc_ping.go .\rpc\messages\rpc_ping.go

%GO_COMPILER_PATH% -r --gen go .\rpc\messages\proto\rpc_request.thrift
copy /y gen-go\rpc\messages\rpc_request.go .\rpc\messages\rpc_request.go

%GO_COMPILER_PATH% -r --gen go .\rpc\messages\proto\rpc_response.thrift
copy /y gen-go\rpc\messages\rpc_response.go .\rpc\messages\rpc_response.go

%GO_COMPILER_PATH% -r --gen go .\rpc\messages\proto\rpc_service.thrift
copy /y gen-go\rpc\messages\rpc_service.go .\rpc\messages\rpc_service.go

@REM network
@REM %GO_COMPILER_PATH% -r --gen go .\network\proto\client_id.thrift
@REM copy /y gen-go\network\client_id.go .\network\client_id.go

@REM cluster
%GO_COMPILER_PATH% -r --gen go .\cluster\proto\prvs\alterrule.thrift
%GO_COMPILER_PATH% -r --gen go .\cluster\proto\prvs\bundle.thrift
%GO_COMPILER_PATH% -r --gen go .\cluster\proto\prvs\closing.thrift
%GO_COMPILER_PATH% -r --gen go .\cluster\proto\pubs\error.thrift
%GO_COMPILER_PATH% -r --gen go .\cluster\proto\pubs\ping.thrift
%GO_COMPILER_PATH% -r --gen go .\cluster\proto\pubs\pubkey.thrift
%GO_COMPILER_PATH% -r --gen go .\cluster\proto\prvs\service.thrift

copy /y gen-go\protocols\prvs\alterrule.go .\cluster\protocols\prvs\alterrule.go
copy /y gen-go\protocols\prvs\bundle.go .\cluster\protocols\prvs\bundle.go
copy /y gen-go\protocols\prvs\closing.go .\cluster\protocols\prvs\closing.go
copy /y gen-go\protocols\prvs\service.go .\cluster\protocols\prvs\service.go

copy /y gen-go\protocols\pubs\error.go .\cluster\protocols\pubs\error.go
copy /y gen-go\protocols\pubs\ping.go .\cluster\protocols\pubs\ping.go
copy /y gen-go\protocols\pubs\pubkey.go .\cluster\protocols\pubs\pubkey.go

@REM %GO_COMPILER_PATH% -r --gen go .\thrift\rpc\messages\rpc_service.thrift


rd/s/q gen-go
echo complate 
pause 
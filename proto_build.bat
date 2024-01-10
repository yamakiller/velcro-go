@echo off 

::协议文件路径, 最后不要跟"\"符号
::set SOURCE_FOLDER=.\parallel\proto

::编译器路径
set GO_COMPILER_PATH=tools\proto\bin\protoc.exe

::删除旧文件
::del .\network\*.pb.go /f /s /q
del .\cluster\protocols\prvs\*.pb.go /f /s /q
del .\cluster\protocols\pubs\*.pb.go /f /s /q
::del .\rpc\messages\*.pb.go /f /s /q

::生成client_id.proto
::echo %GO_COMPILER_PATH% --go_out=.\network --proto_path=.\network\proto client_id.proto
::%GO_COMPILER_PATH% --go_out=.\network --proto_path=.\network\proto client_id.proto

::生成cluster proto intrusive
echo %GO_COMPILER_PATH% --go_out=.\cluster\protocols\prvs --proto_path=.\cluster\proto\prvs --proto_path=.\network\proto closing.proto bundle.proto alterrule.proto logs.proto
%GO_COMPILER_PATH% --go_out=.\cluster\protocols\prvs --proto_path=.\cluster\proto\prvs --proto_path=.\network\proto closing.proto bundle.proto alterrule.proto logs.proto

::生成cluster proto pubs
echo %GO_COMPILER_PATH% --go_out=.\cluster\protocols\pubs --proto_path=.\cluster\proto\pubs --proto_path=.\network\proto error.proto ping.proto pubkey.proto
 %GO_COMPILER_PATH% --go_out=.\cluster\protocols\pubs --proto_path=.\cluster\proto\pubs --proto_path=.\network\proto error.proto ping.proto pubkey.proto

::生成rpc proto
::echo %GO_COMPILER_PATH% --go_out=.\rpc\messages --proto_path=.\rpc\messages\proto --proto_path=.\network\proto rpc_msg.proto rpc_ping.proto rpc_request.proto rpc_response.proto
::%GO_COMPILER_PATH% --go_out=.\rpc\messages --proto_path=.\rpc\messages\proto --proto_path=.\network\proto rpc_msg.proto rpc_ping.proto rpc_request.proto rpc_response.proto

::protoc -I="../actor" --go_out=. --go_opt=paths=source_relative --proto_path=. routercontracts.proto
::pubkey.proto forward_message.proto

echo complate 

pause 
::tools/proto/bin/protoc --go_out=./protocols/ ./protocols/proto/sign.proto
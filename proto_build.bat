@echo off 

::协议文件路径, 最后不要跟"\"符号
::set SOURCE_FOLDER=.\parallel\proto

::编译器路径
set GO_COMPILER_PATH=tools\proto\bin\protoc.exe

::删除旧文件
::del .\network\*.pb.go /f /s /q
del .\cluster\protocols\*.pb.go /f /s /q

::生成client_id.proto
::echo %GO_COMPILER_PATH% --go_out=.\network --proto_path=.\network\proto client_id.proto
::%GO_COMPILER_PATH% --go_out=.\network --proto_path=.\network\proto client_id.proto

::生成cluster proto
echo %GO_COMPILER_PATH% --go_out=.\cluster\protocols --proto_path=.\cluster\proto --proto_path=.\network\proto ping.proto pubkey.proto forward_message.proto
%GO_COMPILER_PATH% --go_out=.\cluster\protocols --proto_path=.\cluster\proto --proto_path=./network/proto ping.proto pubkey.proto forward_message.proto

::protoc -I="../actor" --go_out=. --go_opt=paths=source_relative --proto_path=. routercontracts.proto
::pubkey.proto forward_message.proto

echo complate 

pause 
::tools/proto/bin/protoc --go_out=./protocols/ ./protocols/proto/sign.proto
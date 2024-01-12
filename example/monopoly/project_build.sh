#! /bin/bash

cd .
#::编译器路径
GO_COMPILER_PATH=$(pwd)
echo $GO_COMPILER_PATH

cd $GO_COMPILER_PATH/logs.service 
go build  -o $GO_COMPILER_PATH/bin/logs.service/
cp $GO_COMPILER_PATH/logs.service/config.yaml $GO_COMPILER_PATH/bin/logs.service/config.yaml

cd $GO_COMPILER_PATH/gateway.service 
go build  -o $GO_COMPILER_PATH/bin/gateway.service/
cp $GO_COMPILER_PATH/gateway.service/config.yaml $GO_COMPILER_PATH/bin/gateway.service/config.yaml
cp $GO_COMPILER_PATH/gateway.service/routes.yaml $GO_COMPILER_PATH/bin/gateway.service/routes.yaml

cd $GO_COMPILER_PATH/battle.service 
go build  -o $GO_COMPILER_PATH/bin/battle.service/
cp $GO_COMPILER_PATH/battle.service/config.yaml $GO_COMPILER_PATH/bin/battle.service/config.yaml
cp $GO_COMPILER_PATH/battle.service/routes.yaml $GO_COMPILER_PATH/bin/battle.service/routes.yaml

cd $GO_COMPILER_PATH/login.service 
go build  -o $GO_COMPILER_PATH/bin/login.service/
cp $GO_COMPILER_PATH/login.service/config.yaml $GO_COMPILER_PATH/bin/login.service/config.yaml
cp $GO_COMPILER_PATH/login.service/routes.yaml $GO_COMPILER_PATH/bin/login.service/routes.yaml

#! /bin/bash
cd .

PROJECT_RUN_PATH=$(pwd)
echo $PROJECT_RUN_PATH

$PROJECT_RUN_PATH/logs.service/logs.service install
$PROJECT_RUN_PATH/gateway.service/gateway.service install
$PROJECT_RUN_PATH/login.service/login.service install
$PROJECT_RUN_PATH/battle.service/battle.service install
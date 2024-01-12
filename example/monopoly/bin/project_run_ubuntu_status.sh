#! /bin/bash
cd .

PROJECT_RUN_PATH=$(pwd)
echo $PROJECT_RUN_PATH

$PROJECT_RUN_PATH/logs.service/logs.service status
$PROJECT_RUN_PATH/gateway.service/gateway.service status
$PROJECT_RUN_PATH/login.service/login.service status
$PROJECT_RUN_PATH/battle.service/battle.service status
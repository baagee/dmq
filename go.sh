#!/usr/bin/env bash

basepath=$(
  cd $(dirname $0)
  pwd
)

if [[ $1 == "run" ]]; then
  #  运行
  if [[ $2 == "client" || $2 == "server" ]]; then
    echo "runing..."
    go run $basepath/$2/*.go $basepath/config
  else
    echo "Usage:$0 run server|client"
  fi
elif [[ $1 == "build" ]]; then
  #  生成可执行文件
  if [[ $2 == "client" || $2 == "server" ]]; then
    echo "building..."
    build_file=$basepath/bin/dmq-$2
    rm -rf $build_file
    go build -o $build_file $basepath/$2/*.go
    echo "build over, file: "$build_file
  else
    echo "Usage:$0 build server|client"
  fi
else
  echo "Usage:$0 run|build server|client"
fi

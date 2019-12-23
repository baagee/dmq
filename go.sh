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
    echo "please input server or client"
  fi
elif [[ $1 == "build" ]]; then
  #  生成可执行文件
  if [[ $2 == "client" || $2 == "server" ]]; then
    echo "building..."
    build_file=$basepath/bin/dmq-$2
    rm -rf $build_file
    go build -o $build_file $basepath/$2/*.go
    echo "build over path: "$build_file
  else
    echo "please input server or client"
  fi
else
  echo "please input run or build"
fi

#!/usr/bin/env bash

basepath=$(
  cd $(dirname $0)
  pwd
)

create_dir() {
  if [[ ! -d $1 ]]; then
    mkdir -p -m 755 $1
  fi
}

start() {
  CS=$1
  if [[ $CS == "client" || $CS == "server" ]]; then
    if [[ -f $basepath/var/$CS.pid ]]; then
      #      存在pid 根据pid判断是否正在运行
      pid=$(cat $basepath/var/$CS.pid)
      pid_exists=$(ps aux | awk '{print $2}' | grep -w $pid)
      if [[ $pid_exists ]]; then
        echo "$CS already running"
        return
      fi
    fi
    echo "Start $CS:"
    create_dir $basepath/log
    nohup $basepath/bin/dmq-$CS $basepath/config >$basepath/log/$CS.log &
    pid=$!
    #    创建var目录 保存pid
    create_dir $basepath/var
    echo $pid >$basepath/var/$CS.pid
    echo "ok"
  else
    echo "Usage: $0 start client|server"
  fi
}

stop() {
  CS=$1
  if [[ $CS == "client" || $CS == "server" ]]; then
    if [[ ! -f $basepath/var/$CS.pid ]]; then
      echo "$CS already stop"
      return
    fi
    pid=$(cat $basepath/var/$CS.pid)
    pid_exists=$(ps aux | awk '{print $2}' | grep -w $pid)
    if [[ $pid_exists ]]; then
      echo "Stop $CS:"
      kill $pid
      rm -rf $basepath/var/$CS.pid
      echo "ok"
    fi
  else
    echo "Usage: $0 stop client|server"
  fi
}

case "$1" in
start)
  start $2
  ;;

stop)
  stop $2
  ;;

restart)
  stop $2
  start $2
  ;;
*)
  echo "Usage: $0 start|stop client|server"
  exit 1
  ;;
esac

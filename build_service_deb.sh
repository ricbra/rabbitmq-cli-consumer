#!/bin/bash

set -e

main () {
  read_input
  source ~/.rabbitmq-cli-consumer
  prepare_deb_dirs
  generate_config "$service_name" "$queue_name"
  generate_init "$service_name" "$cmd"
  build_package "$service_name"
  clean_deb_dirs
}

read_input () {
  service_name="rabbitmq-cli-consumer"
  cmd="/var/www/app/console -v acme:cmd"
  queue_name="job_queue"
  if [ -f ~/.rabbitmq-cli-consumer ]; then
    source ~/.rabbitmq-cli-consumer
  fi
  read -e -p "Name of service: " -i "$service_name" service_name
  read -e -p "Command to process queue item: " -i "$cmd" cmd
  read -e -p "Queue name: " -i "$queue_name" queue_name
  cat << EOF > ~/.rabbitmq-cli-consumer
service_name="$service_name"
cmd="$cmd"
queue_name="$queue_name"
EOF
}

prepare_deb_dirs () {
  clean_deb_dirs
  mkdir -p /tmp/deb/etc/init.d
  mkdir -p /tmp/deb/usr/bin
}

generate_config () {
  service_name=$1
  queue_name=$2
  cat <<EOF > /tmp/deb/etc/$service_name.conf
[rabbitmq]
host = localhost
username = guest
password = guest
queue=$queue_name

[logs]
error = /var/log/$service_name.log
info = /var/log/$service_name.log
EOF
}

generate_init () {
  service_name=$1
  cmd=$2
  cat <<EOF > /tmp/deb/etc/init.d/$service_name
#! /bin/sh
### BEGIN INIT INFO
# Provides:          $service_name
# Required-Start:    networking
# Required-Stop:     networking
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: RabbitMQ CLI Consumer $service_name
# Description:       Consumes queues from rabbitmq and executes
#                    command line scripts with message as
#                    argument.
### END INIT INFO

# Using the lsb functions to perform the operations.
. /lib/lsb/init-functions

NAME=$service_name
DAEMON=/usr/bin/\$NAME
PIDFILE=/var/run/\$NAME.pid
CONFIG=/etc/$service_name.conf
CMD='$cmd'

# If the daemon is not there, then exit.
test -x \$DAEMON || exit 5

case \$1 in
 start)
  if [ -e \$PIDFILE ]; then
   status_of_proc -p \$PIDFILE \$DAEMON "\$NAME process" && status="0" || status="\$?"
   if [ \$status = "0" ]; then
    exit # Exit
   fi
  fi
  log_daemon_msg "Starting the process" "\$NAME"
  if start-stop-daemon --start --quiet --oknodo --background --pidfile \$PIDFILE --make-pidfile \\
        --exec \$DAEMON -- --configuration \$CONFIG -V --executable "\$CMD"
  then
   log_end_msg 0
  else
   log_end_msg 1
  fi
  ;;
 stop)
  log_daemon_msg "Stopping \$NAME"
  start-stop-daemon --stop --quiet --oknodo --pidfile \$PIDFILE
  case "\$?" in
          0) log_end_msg 0
             if [ -e \$PIDFILE ]; then
               rm \$PIDFILE
             fi ;;
          1) log_progress_msg "already stopped"
             log_end_msg 0 ;;
          *) log_end_msg 1 ;;
  esac
  ;;
 restart)
  # Restart the daemon.
  \$0 stop && sleep 2 && \$0 start
  ;;
 status)
  # Check the status of the process.
  if [ -e \$PIDFILE ]; then
   status_of_proc -p \$PIDFILE \$DAEMON "\$NAME process" && exit 0 || exit \$?
  else
   log_daemon_msg "\$NAME process is not running"
   log_end_msg 0
  fi
  ;;
 *)
  # For invalid arguments, print the usage message.
  echo "Usage: \$0 {start|stop|restart|status}"
  exit 2
  ;;
esac
EOF
  chmod +x /tmp/deb/etc/init.d/$service_name
}

build_package () {
  service_name=$1
  if [ ! -d ~/gocode ]; then
    mkdir ~/gocode
  fi
  export GOPATH=~/gocode
  go get github.com/ricbra/rabbitmq-cli-consumer
  go build github.com/ricbra/rabbitmq-cli-consumer
  cp $GOPATH/src/github.com/ricbra/rabbitmq-cli-consumer/rabbitmq-cli-consumer /tmp/deb/usr/bin/$service_name
  VERSION=`/tmp/deb/usr/bin/$service_name --version|awk '{print $3}'`
  echo "fpm -s dir -t deb -C /tmp/deb --force --name $service_name --version $VERSION --description \"Consumes RabbitMQ messages into cli program\" --config-files etc/$service_name.conf"
}

clean_deb_dirs () {
  rm -rf /tmp/deb
}

main

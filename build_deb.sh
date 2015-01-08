#!/bin/bash

ARCHS['amd64']='amd64'
ARCHS['386']='i386'

DIR=`mktemp -d`

for ARCH in 'amd64' '386'
do
    BUILD=${ARCHS[$ARCH]}
    gox -osarch="linux/$ARCH" -output="$DIR/$ARCH/usr/bin/{{.Dir}}"
    fpm -s dir -t deb -C $DIR/$ARCH -a $BUILD --name rabbitmq-cli-consumer --version $1 --description "Consume RabbitMQ messages into any cli program" .
done

rm -rf $DIR



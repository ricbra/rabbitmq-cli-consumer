#!/bin/bash

if [ -z ${1+x} ];
    then
        echo "Missing version number!";
        exit
    else
        echo "Preparing version $1"
fi

ARCHS['amd64']='amd64'
ARCHS['386']='i386'

DIR=`mktemp -d /tmp/rabbitmq.XXXX`
CWD=`pwd`

for ARCH in 'amd64' '386'
do
    BUILD=${ARCHS[$ARCH]}
    gox -osarch="linux/$ARCH" -output="$DIR/$ARCH/usr/bin/{{.Dir}}"
    fpm -s dir -t deb -C $DIR/$ARCH -a $BUILD --name rabbitmq-cli-consumer --version $1 --description "Consume RabbitMQ messages into any cli program" .
done

for ARCH in 'linux/amd64' 'linux/386' 'linux/arm' 'darwin/amd64'
do
    arch_suffix=${ARCH/\//-}
    gox -osarch="$ARCH" -output="$DIR/{{.Dir}}"
    cd $DIR
    tar -zcvf $CWD/rabbitmq-cli-consumer-$arch_suffix.tar.gz rabbitmq-cli-consumer
    rm -rf rabbitmq-cli-consumer
    cd $CWD
done

rm -rf $DIR

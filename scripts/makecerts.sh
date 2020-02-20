#!/bin/bash
# ./makecert.sh certs joe@random.com
#               ^ dir for certs
#                     ^ email
DIR=certs
mkdir ${DIR}
rm ${DIR}/*
echo "make server cert"
openssl req -new -nodes -x509 -out ${DIR}/server.pem -keyout certs/server.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=www.random.com/emailAddress=$2"
echo "make client cert"
openssl req -new -nodes -x509 -out ${DIR}/client.pem -keyout certs/client.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=www.random.com/emailAddress=$2"
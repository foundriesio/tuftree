#!/bin/sh -e

HERE=$(readlink -f $(dirname 0))
cd $HERE

openssl ecparam -genkey -name prime256v1 -noout -out root-ca.key
openssl req -x509 -new -key root-ca.key -out ./fixtures/root-ca.crt -subj "/CN=Notary Intermediate CA"

# Generate notary-server key with 10 year validity. It binds to notary-server:
openssl ecparam -genkey -name prime256v1 -noout -out ./fixtures/notary-server.key
openssl req -new -key fixtures/notary-server.key -out server.csr -subj "/CN=tuftree-notary"
openssl x509 -req -in server.csr -CAcreateserial \
	-CAkey root-ca.key -CA fixtures/root-ca.crt -out fixtures/notary-server.crt -days 3650

openssl ecparam -genkey -name prime256v1 -noout -out fixtures/notary-signer.key
openssl req -new -key fixtures/notary-signer.key -out signer.csr -subj "/CN=notary-signer"
openssl x509 -req -in signer.csr -CAcreateserial \
	-CAkey root-ca.key -CA fixtures/root-ca.crt -out fixtures/notary-signer.crt -days 3650

rm root-ca.key *.csr fixtures/root-ca.srl
chmod go+r fixtures/*

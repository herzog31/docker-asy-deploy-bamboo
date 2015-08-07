FROM miek/alpine-armv6l

MAINTAINER Mark J. Becker <mjb@marb.ec>

RUN apk --update add openssl ca-certificates unzip curl py-pip py-yaml
RUN pip install -U docker-compose

RUN mkdir /deploy-root
RUN chmod -R 0777 /deploy-root
WORKDIR /deploy-root

COPY asy-deploy /deploy-root/
CMD ["./asy-deploy"]
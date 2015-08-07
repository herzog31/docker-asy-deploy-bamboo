FROM miek/alpine-armv6l

RUN apk --update add openssl ca-certificates unzip curl py-pip py-yaml
# RUN curl -L https://github.com/hypriot/compose/releases/download/1.2.0-raspbian/docker-compose-Linux-armv6l > /usr/bin/docker-compose
# RUN chmod +x /usr/bin/docker-compose
RUN pip install -U docker-compose

RUN mkdir /deploy-root
RUN chmod -R 0777 /deploy-root
WORKDIR /deploy-root

COPY asy-deploy /deploy-root/
CMD ["./asy-deploy"]
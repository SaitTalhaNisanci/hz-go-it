FROM golang:1.9.2

RUN mkdir -p /local/git
WORKDIR /local/git

ADD client.sh client.sh
RUN chmod +x ./*.sh

CMD ["/bin/sh", "-c", "./client.sh"]
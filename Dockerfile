FROM openjdk:8-slim

# update packages and install golang
RUN  \
  export DEBIAN_FRONTEND=noninteractive && \
  sed -i 's/# \(.*multiverse$\)/\1/g' /etc/apt/sources.list && \
  apt-get update && \
  apt-get -y upgrade && \
  apt-get install -y vim wget curl git golang

# create working directory
RUN mkdir -p /local/git
WORKDIR /local/git

ADD client.sh client.sh
RUN chmod +x ./*.sh

CMD ["/bin/sh", "-c", "./client.sh"]
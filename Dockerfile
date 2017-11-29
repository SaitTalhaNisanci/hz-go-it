FROM golang:1.9.2

#CMD \
#    go get github.com/docker/libcompose/docker && \
#    go get github.com/docker/libcompose/docker/ctx && \
#    go get github.com/docker/libcompose/project && \
#    go get github.com/docker/libcompose/project/options && \
#    go get github.com/stretchr/testify/assert && \
#    go get github.com/lucasjones/reggen && \
#    go get github.com/montanaflynn/stats
#
#RUN echo "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
#RUN mkdir /root/.ssh && echo "StrictHostKeyChecking no " > /root/.ssh/config
#
#CMD \
#    go get github.com/lazerion/go-client && \
#    go get github.com/lazerion/go-client/config && \
#    go get github.com/lazerion/go-client/core

# create working directory
RUN mkdir -p /local/git
WORKDIR /local/git

ADD client.sh client.sh
RUN chmod +x ./*.sh

CMD ["/bin/sh", "-c", "./client.sh"]
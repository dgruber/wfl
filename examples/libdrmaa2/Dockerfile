FROM drmaa/gridengine

RUN yum install -y wget tar git gcc

RUN export VERSION=1.16 OS=linux ARCH=amd64 && \
  wget https://dl.google.com/go/go$VERSION.$OS-$ARCH.tar.gz && \
  tar -C /usr/local -xzvf go$VERSION.$OS-$ARCH.tar.gz && \
  rm go$VERSION.$OS-$ARCH.tar.gz

ENV GOPATH /go
ENV PATH /usr/local/go/bin:${PATH}:${GOPATH}/bin
ENV PATH ${PATH}:/opt/sge/include

ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]

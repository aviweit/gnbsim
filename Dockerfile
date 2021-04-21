FROM golang:1.15.6-buster 

RUN apt-get update && apt-get install -y tcpdump vim strace net-tools curl netcat-openbsd iptables python python-pip python-setuptools python3 python3-pip 

RUN mkdir /go/src/gnbsim

ADD . /go/src/gnbsim

RUN cd /go/src/gnbsim &&\
    make

WORKDIR /go/src/gnbsim/gnb-gw
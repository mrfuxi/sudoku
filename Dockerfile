FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get install -y software-properties-common
RUN add-apt-repository -y ppa:amarburg/opencv3
RUN apt-get update
RUN apt-get install libopencv3

FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get install -y software-properties-common
RUN add-apt-repository -y ppa:amarburg/opencv3
RUN apt-get update
RUN apt-get install -y python-opencv3 python-pip python-dev python-zmq
RUN pip install ipython[notebook]

EXPOSE 8888

WORKDIR /code

CMD ipython notebook --ip=0.0.0.0 --no-browser

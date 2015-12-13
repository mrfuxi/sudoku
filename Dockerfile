FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get install -y software-properties-common
RUN add-apt-repository -y ppa:amarburg/opencv3
RUN apt-get update
RUN apt-get install -y python-opencv3 python-pip python-dev
RUN apt-get install -y tesseract-ocr-eng tesseract-ocr python-pil

RUN echo "tessedit_char_whitelist 123456789" > /usr/share/tesseract-ocr/tessdata/configs/sudoku

RUN mkdir /code
WORKDIR /code
ADD requirements.txt requirements.txt
RUN pip install -r requirements.txt
ADD . ./
VOLUME /code

EXPOSE 8888

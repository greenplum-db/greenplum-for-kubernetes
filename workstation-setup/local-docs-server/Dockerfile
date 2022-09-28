FROM ubuntu:latest

RUN apt-get update && \
      apt-get install -y \
        ruby-full \
        git \
        vim \
        build-essential \
        patch \
        ruby-dev \
        zlib1g-dev \
        liblzma-dev

RUN gem install bundler && \
        gem update --system


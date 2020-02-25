#!/bin/bash

VERSION="v1.0.2"

USRDIR="/usr/local/cloudcare/forethought/wdf"
CONF="${USRDIR}/wdf.conf"
HOST_DIR="/cloudcare-wdf"
NSQD_DATA_DIR="${HOST_DIR}/nsqd-data"

WDF_PORT="8080"

NSQ_LOOKUPD_HOST=$1

info() {
  printf "[$(date +'%Y-%m-%d %H:%M:%S')] \033[32m INFO\033[0m $1\n"
}

err() {
  printf "[$(date +'%Y-%m-%d %H:%M:%S')] \033[31mERROR\033[0m $1\n"
}

function command_check() {

  if [ -x "$(command -v go)" ]; then
    info "Command go ok"
  else 
    err "Command go is not found"
    return 1
  fi

  if [ -x "$(command -v docker)" ]; then
    info "Command docker ok"
  else 
    err "Command docker is not found"
    return 1
  fi

  if [ -x "$(command -v docker-compose)" ]; then
    info "Command docker-compose ok"
  else 
    err "Command docker-compose is not found"
    return 1
  fi

  return 0
}

function creata_files() {

  if [ ! -d ${NSQD_DATA_DIR} ]; then
    if ( sudo mkdir -p ${NSQD_DATA_DIR} ) ; then
      info "Create ${NSQD_DATA_DIR} success"
    else 
      err "Create ${NSQD_DATA_DIR} failed"
      exit 1
    fi
  fi

  return 0
}

function docker_build_wdf() {

  dockerfile="# wdf build image

FROM ubuntu

RUN mkdir -p ${USRDIR}

CMD ${USRDIR}/wdf -cfg ${CONF}
"
  echo "${dockerfile}" > Dockerfile
  info "Dockerfile writing.."

  if ( go build -o wdf main.go ) ; then
    info "Build wdf binary success"
  else
    info "Build wdf binary failed"
    return 1
  fi

  if ( sudo cp wdf ${HOST_DIR} && \
       sudo cp wdf.conf.example "${HOST_DIR}/wdf.conf" )  ; then
      info "Copy wdf files to ${HOST_DIR} success"
  else
      err "Copy wdf files to ${HOST_DIR} failed"
      exit 1
  fi

  if ( sudo docker build -t wdf:${VERSION} . ) ; then
    info "Build wdf docker image wdf:${VERSION}.."
  else
    info "Build wdf docker image failed"
    return 1
  fi

  return 0
}

function docker_install_NSQ() {

  docker_compose_yml="version: '3'

services:
  nsqlookup:
    image: nsqio/nsq
    hostname: nsqlookup
    ports:
      - \"4160:4160\"    # tcp port
      - \"4161:4161\"    # http port
    command: /nsqlookupd 

  nsq:
    image: nsqio/nsq
    hostname: nsq
    ports:
      - \"4150:4150\"    # tcp port， 这两个端口需要相等
      - \"4151:4151\"    # http port，这两个端口需要相等
    links:
       - nsqlookup:nsqlookup
    volumes:
       - ${NSQD_DATA_DIR}:/data/nsqd
    command: /nsqd --broadcast-address \"${NSQ_LOOKUPD_HOST}\" --lookupd-tcp-address=nsqlookup:4160 --data-path=/data/nsqd --max-msg-timeout 30h --max-msg-size 1048576

  nsqadmin:
    image: nsqio/nsq
    hostname: nsqadmin
    links:
      - nsqlookup:nsqlookup
    ports:
      - \"4171:4171\"
    command: /nsqadmin --lookupd-http-address=nsqlookup:4161

  wdf:
    image: wdf:${VERSION}
    hostname: "wdf-${VERSION}" 
    ports:
      - "${WDF_PORT}":8080
    links:
       - nsqlookup:nsqlookup
    volumes:
      - ${HOST_DIR}:"${USRDIR}"
    command: ${USRDIR}/wdf -cfg ${USRDIR}/wdf.conf
"
  echo "${docker_compose_yml}" > docker-compose.yml
  info "Docker-compose.yaml writing.."

  if ( sudo docker-compose up -d ); then
    info "docker-compose up success"
  else
    info "docker-compose up failed"
    return 1
  fi

  return 0
}

if command_check 0 ; then
  info "Check command is ok"
else
  err "Check command is failed"
  exit 1
fi

if docker_build_wdf 0 ; then
  info "Docker build wdf success"
else
  err "Docker build wdf failed"
  exit 1
fi

if docker_install_NSQ 0 ; then
  info "Docker install NSQ success"
else
  err "Docker install NSQ failed"
  exit 1
fi

info "Docker wdf run success, http port :8080"
info "Nsq consumer host is ${NSQ_LOOKUPD_HOST}:14161"
info "End.."

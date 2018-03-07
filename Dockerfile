FROM golang:1.8

# Megalith configuration
ENV WEBAPP /go/src/server
ENV PORT=9001
# Hostname of dispatcher server/cluster 
ENV DISPATCHER_HOSTNAME megalith-dispatcher
ENV WORKER_HOSTNAME megalith
ENV DISPATCHER_ADDR=http://${DISPATCHER_HOSTNAME}:${PORT}
ENV WORKER_ADDR=http://${WORKER_HOSTNAME}:${PORT}


RUN mkdir -p ${WEBAPP}
COPY . ${WEBAPP}
WORKDIR ${WEBAPP}
RUN go install
RUN rm -rf ${WEBAPP} 
EXPOSE $PORT
# Comment/uncomment the following line to
# disable/enable worker mode
CMD server -worker -container

# Comment/uncomment the following line to
# disable/enable dispatcher mode.
# build with `docker build -t megalithDispatcher .`
# CMD server -nobrowser -container
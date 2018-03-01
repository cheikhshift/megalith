FROM golang:1.8
ENV WEBAPP /go/src/server
RUN mkdir -p ${WEBAPP}
COPY . ${WEBAPP}
ENV PORT=9001 
WORKDIR ${WEBAPP}
RUN go install
RUN rm -rf ${WEBAPP} 
EXPOSE 9001
CMD server

FROM golang:1.8
RUN mkdir -p /go/src/server
RUN mkdir -p /var/pool
COPY . /go/src/server/
ENV PORT=9001 
RUN cd /go/src/server && go install
EXPOSE 9001
EXPOSE 8700
CMD cd /go/src/server && server
# healthcheck requires docker 1.12 and up.
# HEALTHCHECK --interval=20m --timeout=3s \
#  CMD curl -f http://localhost:9001/ || exit 1
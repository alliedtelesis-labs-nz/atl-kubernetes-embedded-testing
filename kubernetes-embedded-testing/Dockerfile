FROM golang:1.24

COPY . .

RUN make build \
  && ls -l ./bin \
  && cp ./bin/ket /usr/local/bin/ket
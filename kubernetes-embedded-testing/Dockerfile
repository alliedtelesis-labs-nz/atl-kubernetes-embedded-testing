FROM golang:1.24

copy . .

RUN make build && \
  cp ./bin/ket /usr/local/bin/ket && \
  rm -rf .
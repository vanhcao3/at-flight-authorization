FROM harbor.vht.vn/proxy-cache/golang:1.24 AS builder
LABEL stage=builder

ARG SVC

WORKDIR /go/src/go-template
COPY . .

RUN make go-build \
    && mv bin/$SVC /exe

FROM scratch
COPY --from=builder /exe /
ENTRYPOINT ["./exe", "start"]

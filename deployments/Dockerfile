FROM --platform=$TARGETPLATFORM golang:1.15 as builder

WORKDIR /vanga/

#Install deps and build
ENV GO111MODULE=on

ENV GOPROXY=http://goproxy.tcfw.com.au,direct

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN make _SERVICE_

# ---
FROM --platform=$TARGETPLATFORM gruebel/upx:latest as upx

COPY --from=builder /vanga/build/_SERVICE_ /_SERVICE_.org
RUN upx --best --lzma -o /_SERVICE_ /_SERVICE_.org

# ---

FROM --platform=$TARGETPLATFORM ubuntu:latest

RUN apt update && apt install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /bin
COPY --from=upx /_SERVICE_ _SERVICE_

CMD [ "/bin/_SERVICE_", "serve" ]
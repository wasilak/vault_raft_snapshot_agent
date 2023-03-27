FROM quay.io/wasilak/golang:1.20-alpine AS builder

RUN mkdir /build
WORKDIR /build

COPY . .

RUN go mod download
RUN go build -o ../vault_raft_snapshot_agent .

FROM quay.io/wasilak/alpine:3
WORKDIR /
COPY --from=builder /vault_raft_snapshot_agent .
COPY snapshot.json /snapshot.json
ENTRYPOINT ["/vault_raft_snapshot_agent"]

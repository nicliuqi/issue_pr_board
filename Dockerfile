FROM golang:latest as BUILDER
MAINTAINER liuqi<469227928@qq.com>
COPY . /go/src/github.com/opensourceways/issue_pr_board
RUN cd /go/src/github.com/opensourceways/issue_pr_board && go mod tidy && CGO_ENABLED=1 go build -v -o ./ipb main.go sync.go

FROM openeuler/go:1.17.3-22.03-lts
COPY ./conf /opt/app/conf
COPY --from=BUILDER /go/src/github.com/opensourceways/issue_pr_board/ipb /opt/app
WORKDIR /opt/app/
ENTRYPOINT ["/opt/app/ipb"]

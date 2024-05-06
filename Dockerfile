FROM openeuler/openeuler:23.09 as BUILDER
MAINTAINER liuqi<469227928@qq.com>
RUN sed -i "s|repo.openeuler.org|mirrors.nju.edu.cn/openeuler|g" /etc/yum.repos.d/openEuler.repo \
 && sed -i "/metalink/d" /etc/yum.repos.d/openEuler.repo \
 && sed -i "/metadata_expire/d" /etc/yum.repos.d/openEuler.repo \
 && yum install -y golang
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY . /go/src/github.com/opensourceways/issue_pr_board
RUN cd /go/src/github.com/opensourceways/issue_pr_board && go mod tidy && CGO_ENABLED=1 go build -v -o ./ipb main.go sync.go

FROM openeuler/openeuler:22.03
ARG user=ipb
ARG group=ipb
ARG uid=1000
ARG gid=1000

RUN groupadd -g ${gid} ${group}
RUN useradd -u ${uid} -g ${group} -d /home/ipb/ -s /sbin/nologin -m ${user}
COPY ./conf /home/ipb/conf
COPY ./utils/aj-captcha/resources /home/ipb/resources
COPY ./templates /home/ipb/templates
COPY --from=BUILDER /go/src/github.com/opensourceways/issue_pr_board/ipb /home/ipb
WORKDIR /home/ipb/
RUN chown -R ${user}:${group} /home/ipb
USER ${uid}:${gid}
ENV TZ=Asia/Shanghai
ENTRYPOINT ["/home/ipb/ipb"]

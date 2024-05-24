FROM openeuler/openeuler:23.09 as BUILDER
MAINTAINER liuqi<469227928@qq.com>
RUN sed -i "s|repo.openeuler.org|mirrors.nju.edu.cn/openeuler|g" /etc/yum.repos.d/openEuler.repo \
 && sed -i "/metalink/d" /etc/yum.repos.d/openEuler.repo \
 && sed -i "/metadata_expire/d" /etc/yum.repos.d/openEuler.repo \
 && yum install -y golang
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY . /go/src/github.com/opensourceways/issue_pr_board
RUN cd /go/src/github.com/opensourceways/issue_pr_board && go mod tidy && GO111MODULE=on CGO_ENABLED=0 go build -v -o ./ipb main.go sync.go

FROM openeuler/openeuler:22.03
ARG user=ipb
ARG group=ipb
ARG uid=1000
ARG gid=1000

RUN groupadd -g ${gid} ${group}
RUN useradd -u ${uid} -g ${group} -d /home/ipb/ -s /sbin/nologin -m ${user}
RUN rm -rf /usr/bin/kill
RUN echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd
RUN sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs
RUN echo 'set +o history' >> /root/.bashrc
RUN rm -rf /tmp/*

COPY ./conf /home/ipb/conf
COPY ./templates /home/ipb/templates
COPY --from=BUILDER /go/src/github.com/opensourceways/issue_pr_board/ipb /home/ipb
RUN chmod -R 550 /home/ipb
RUN chmod +x /home/ipb/ipb
RUN mkdir -p /home/ipb/logs
RUN chmod -R 750 /home/ipb/logs
WORKDIR /home/ipb/
RUN chown -R ${user}:${group} /home/ipb
RUN history -c && echo "set +o history" >> /home/ipb/.bashrc  && echo "umask 027" >> /home/ipb/.bashrc && source /home/ipb/.bashrc
ENV TZ=Asia/Shanghai
USER ${uid}:${gid}
ENTRYPOINT ["/home/ipb/ipb"]

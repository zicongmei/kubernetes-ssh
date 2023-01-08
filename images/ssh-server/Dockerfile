FROM ubuntu:20.04

RUN apt update
RUN DEBIAN_FRONTEND=noninteractive apt install -y openssh-server
RUN apt install -y nfs-kernel-server nfs-common
RUN mkdir -p /run/sshd

ADD ssh_config /etc/ssh
ADD sshd_config /etc/ssh

CMD /usr/sbin/sshd -D

EXPOSE 22
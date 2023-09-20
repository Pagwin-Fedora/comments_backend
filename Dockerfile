FROM golang AS build

ADD . /root/
WORKDIR /root
RUN go build

FROM alpine as certs
RUN apk update && apk add ca-certificates

FROM busybox

# making a user to run the app so if my shitty code has remote code exec it's less problematic
RUN ["/bin/mkdir", "-p", "/application"]
RUN ["/bin/adduser", "-h", "/application", "-D", "application"]
RUN ["/bin/chown", "-R", "application:application", "/application"]
USER application

# port used
EXPOSE 80/tcp
# default is localhost due to assuming it being run outside of a container but due to dockern networking we need to change it
ENV DB_URI="host.docker.internal"

# templates
ADD email_verification.html /application/email_verification.html
ADD email_sent_notif.html /application/email_sent_notif.html
ADD comment.html /application/comment.html

# adding TLS certificates so mailjet can work
COPY --from=certs /etc/ssl/certs /etc/ssl/certs

# adding in the actual executable
COPY --from=build --chown=application:application /root/comments /application/comments
RUN ["/bin/chmod", "+x", "/application/comments"]

WORKDIR /application
CMD ["/application/comments"]

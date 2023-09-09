FROM golang AS build

ADD . /root/
WORKDIR /root
RUN go build

FROM busybox

# making a user to run the app because why not
RUN ["/bin/mkdir", "-p", "/application"]
RUN ["/bin/adduser", "-h", "/application", "-D", "application"]
RUN ["/bin/chown", "application:application", "/application"]

EXPOSE 8080/tcp
# default is localhost due to assuming it being run outside of a container but due to dockern networking we need to change it
ENV DB_URI="host.docker.internal"
COPY --from=build --chown=application:application /root/comments /application/comments
CMD ["/application/comments"]

FROM golang AS build

ADD . /root/
WORKDIR /root
RUN go build

FROM busybox

# making a user to run the app because why not
RUN ["/bin/mkdir", "-p", "/application"]
RUN ["/bin/adduser", "-h", "/application", "-D", "application"]
RUN ["/bin/chown", "-R", "application:application", "/application"]
USER application

EXPOSE 80/tcp
# default is localhost due to assuming it being run outside of a container but due to dockern networking we need to change it
ENV DB_URI="host.docker.internal"
COPY --from=build --chown=application:application /root/comments /application/comments
#COPY --from=build --chown=application:application /root/index.html /application/index.html
RUN ["/bin/chmod", "+x", "/application/comments"]
WORKDIR /application
CMD ["/application/comments"]

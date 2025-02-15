FROM debian


RUN apt-get update && apt-get install -y \
    glusterfs-server

RUN mkdir /build

COPY ./main /build/glusterfs-csi-driver

ENTRYPOINT ["/build/glusterfs-csi-driver"]


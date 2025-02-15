# Copyright 2018 The Gluster CSI Authors.

# Licensed under GNU LESSER GENERAL PUBLIC LICENSE Version 3, 29 June 2007
# You may obtain a copy of the License at
#    https://opensource.org/licenses/lgpl-3.0.html

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#-- Create build environment

FROM golang as build

ENV GOPATH="/go/"

ENV SRCDIR="/go/src/github.com/gluster/gluster-csi-driver"
ENV SCRIPTSDIR="${SRCDIR}/scripts"

RUN apt-get install -y \
    curl \
    git

# Install go tools
ARG GO_METALINTER_VERSION=latest
WORKDIR "${SRCDIR}"
COPY scripts/install-go-tools.sh "${SRCDIR}"
RUN mkdir -p /go/binary; ${SRCDIR}/install-go-tools.sh

# Vendor dependencies
COPY go.mod "${SRCDIR}"
COPY go.sum "${SRCDIR}"
RUN go get; go mod tidy

# Copy source directories
COPY cmd/ "${SRCDIR}/cmd"
COPY pkg/ "${SRCDIR}/pkg"
COPY scripts/ "${SCRIPTSDIR}"


RUN mkdir /build

#-- Build phase

# Build executable
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o /build/glusterfs-csi-driver cmd/glusterfs/main.go

# # Ensure the binary is statically linked
RUN ldd /build/glusterfs-csi-driver | grep -q "not a dynamic executable"

# #-- Final container

FROM ubuntu as final

# Install dependencies
## Use my release
# ADD http://artifacts.ci.centos.org/gluster/nightly/release-8.repo  /etc/yum.repos.d/glusterfs-nightly.repo
RUN apt-get update -y && \
    apt -y install glusterfs-server
    # yum clean all -y && \
    # rm -rf /var/cache/yum && \
    # rpm -qa | grep gluster | tee /gluster-rpm-versions.txt

# Copy glusterfs-csi-driver from build phase
COPY --from=build /build /

# The version of the driver (gUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o /build/glusterfs-csi-driver cmd/glusterfs/main.go

# # Ensure the binary is statically linked
RUN ldd /build/glusterfs-csi-driver | grep -q "not a dynamic executable"


# Install dependencies
## Use my releasit describe --dirty --always --tags | sed 's/-/./2' | sed 's/-/./2')
ARG version="(unknown)"
# Container build time (date -u '+%Y-%m-%dT%H:%M:%S.%NZ')
ARG builddate="(unknown)"

LABEL build-date="${builddate}"
LABEL io.k8s.description="FUSE-based CSI driver for Gluster file access"
LABEL name="glusterfs-csi-driver"
LABEL Summary="FUSE-based CSI driver for Gluster file access"
LABEL vcs-type="git"
LABEL vcs-url="https://github.com/gluster/gluster-csi-driver"
LABEL vendor="gluster.org"
LABEL version="${version}"

ENTRYPOINT ["/glusterfs-csi-driver"]

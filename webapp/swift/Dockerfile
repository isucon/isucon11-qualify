FROM fabianfett/amazonlinux-swift:5.2-amazonlinux2
ENV SERVICE_NAME=EmptyExampleService
WORKDIR /swift/application/src/
COPY Sources ./Sources/
COPY Tests ./Tests/
COPY Package.* ./
RUN yum install -y git gcc gcc-c++ aws-cli jq tar wget which iptables openssl-devel zlib-devel make libedit-devel
RUN swift build -c release --build-path .build/native --disable-prefetching 
RUN swift test
RUN mkdir -p .build/service/libraries
RUN ldd .build/native/release/${SERVICE_NAME} | grep '=>' | sed -e '/^[^\t]/ d' | sed -e 's/\t//' | sed -e 's/.*=..//' | sed -e 's/ (0.*)//' | xargs -i% cp % .build/service/libraries
RUN cp .build/native/release/${SERVICE_NAME} .build/service/
RUN ls .build/service

FROM amazonlinux:2  
RUN mkdir app
WORKDIR /app
COPY --from=0 /swift/application/src/.build/service /app/
EXPOSE 8080
CMD ["/lib64/ld-linux-x86-64.so.2", "--library-path", "libraries", "./EmptyExampleService"]

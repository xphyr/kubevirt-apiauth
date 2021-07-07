FROM registry.access.redhat.com/ubi8/ubi-minimal AS build

RUN microdnf install golang
WORKDIR /src/
COPY main.go go.* /src/
RUN go build -o /bin/listvms

FROM registry.access.redhat.com/ubi8/ubi-minimal
COPY --from=build /bin/listvms /bin/listvms
ENTRYPOINT ["/bin/listvms"]

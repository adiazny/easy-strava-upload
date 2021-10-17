FROM alpine:latest as base

FROM scratch

COPY easy-strava-upload .

USER 65534

ENTRYPOINT ["./easy-strava-upload"]

FROM alpine:edge AS build
RUN apk update
RUN apk upgrade
RUN apk add --update go=1.12.7-r0 gcc=8.3.0-r0 g++=8.3.0-r0 git
WORKDIR /app
COPY . . 

RUN CGO_ENABLED=1 GO111MODULE=on GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest  
WORKDIR /root/
COPY --from=0 /app/Traefik-U2F-SSO .
CMD ["./Traefik-U2F-SSO"]  

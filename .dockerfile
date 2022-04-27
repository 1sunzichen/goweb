FROM goweb:latest
RUN mkdir -p /app
WORKDIR /app

ADD main /app/main

EXPOSE 8080

CMD ["./main"]
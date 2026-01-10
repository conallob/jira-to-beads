FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY jira-to-beads /usr/local/bin/jira-to-beads

ENTRYPOINT ["jira-to-beads"]
CMD ["help"]

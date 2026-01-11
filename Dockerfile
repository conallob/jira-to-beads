FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY jira-beads-sync /usr/local/bin/jira-beads-sync

ENTRYPOINT ["jira-beads-sync"]
CMD ["help"]

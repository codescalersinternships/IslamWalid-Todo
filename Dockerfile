FROM golang

WORKDIR /todo_api

COPY . .

RUN go mod download

EXPOSE 8080

CMD ["go", "run", "todo.go"]

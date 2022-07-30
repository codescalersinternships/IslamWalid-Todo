FROM golang

WORKDIR /todo_api

COPY . .

RUN go mod download

CMD ["go", "run", "todo.go"]

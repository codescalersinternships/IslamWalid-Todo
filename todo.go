package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)


const (
    JsonIndent = "    "
    Port = ":8080"

    IDExists = "Task with ID %d already exists\n"
    IDNotFound = "Task with ID %d was not found\n"

    dbFile = "./database/todo.db"
)

var db *gorm.DB

type Task struct {
    ID uint64       `json:"id"`
    Title string    `json:"title"`
    Completed bool  `json:"completed"`
}

func main() {
    router := mux.NewRouter()
    InitDB()
    RegisterHandlers(router)
    http.ListenAndServe(Port, router)
}

func InitDB() {
    db, _ = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
    db.AutoMigrate(&Task{})
}

func RegisterHandlers(router *mux.Router) {
    router.HandleFunc("/todo", GetAllTasks).Methods("GET")
    router.HandleFunc("/todo", AddTask).Methods("POST")
    router.HandleFunc("/todo/{id}", GetTaskByID).Methods("GET")
}

var GetAllTasks = func (writer http.ResponseWriter, req *http.Request) {
    tasks := make([]Task, 0)

    db.Find(&tasks)
    tasksJson, _ := json.MarshalIndent(tasks, "", JsonIndent)
    httpResponse(writer, tasksJson, http.StatusOK)
}

var AddTask = func (w http.ResponseWriter, req *http.Request) {
    var newTask Task
 
    json.NewDecoder(req.Body).Decode(&newTask)

    if db.First(&Task{}, newTask.ID).Error == nil {
        http.Error(w, fmt.Sprintf(IDExists, newTask.ID), http.StatusConflict)
    } else {
        db.Create(&newTask)
        newTaskJson, _ := json.MarshalIndent(newTask, "", JsonIndent)
        httpResponse(w, newTaskJson, http.StatusCreated)
    }
}

var GetTaskByID = func (w http.ResponseWriter, req *http.Request) {
    var resultTask Task

    params := mux.Vars(req)
    id, _ := strconv.Atoi(params["id"])

    if db.First(&resultTask, id).Error == nil {
        resultJson, _ := json.MarshalIndent(resultTask, "", JsonIndent)
        httpResponse(w, resultJson, http.StatusOK)
    } else {
        http.Error(w, fmt.Sprintf(IDNotFound, id), http.StatusNotFound)
    }
}

func httpResponse(w http.ResponseWriter, data []byte, statusCode int)  {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    w.Write(data)
}

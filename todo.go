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
    BadRequest = "Request is not valid\n"

    dbFile = "./database/todo.db"
)

var db *gorm.DB

type Task struct {
    ID int          `json:"id"`
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
    router.HandleFunc("/todo", ModifyTask).Methods("PATCH")
    router.HandleFunc("/todo/{id}", GetTaskByID).Methods("GET")
    router.HandleFunc("/todo/{id}", DeleteTaskByID).Methods("DELETE")
}

var GetAllTasks = func (w http.ResponseWriter, req *http.Request) {
    tasks := make([]Task, 0)

    db.Find(&tasks)
    tasksJson, _ := json.MarshalIndent(tasks, "", JsonIndent)
    httpResponse(w, tasksJson, http.StatusOK)
}

var AddTask = func (w http.ResponseWriter, req *http.Request) {
    var newTask Task
 
    err := json.NewDecoder(req.Body).Decode(&newTask)
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    _, isExist := DbIsExist(newTask.ID)
    if isExist {
        http.Error(w, fmt.Sprintf(IDExists, newTask.ID), http.StatusConflict)
    } else {
        db.Create(&newTask)
        newTaskJson, _ := json.MarshalIndent(newTask, "", JsonIndent)
        httpResponse(w, newTaskJson, http.StatusCreated)
    }
}

var GetTaskByID = func (w http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    task, isExist := DbIsExist(id)
    if isExist {
        resultJson, _ := json.MarshalIndent(task, "", JsonIndent)
        httpResponse(w, resultJson, http.StatusOK)
    } else {
        http.Error(w, fmt.Sprintf(IDNotFound, id), http.StatusNotFound)
    }
}

var ModifyTask = func (w http.ResponseWriter, req *http.Request) {
    var newTask Task

    err := json.NewDecoder(req.Body).Decode(&newTask)
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    task, isExist := DbIsExist(newTask.ID)
    if isExist {
        task.Title = newTask.Title
        task.Completed = newTask.Completed
        db.Save(task)
        resultJson, _ := json.MarshalIndent(task, "", JsonIndent)
        httpResponse(w, resultJson, http.StatusOK)
    } else {
        http.Error(w, fmt.Sprintf(IDNotFound, newTask.ID), http.StatusNotFound)
    }
}

var DeleteTaskByID = func (w http.ResponseWriter, req *http.Request)  {
    params := mux.Vars(req)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    task, isExist := DbIsExist(id)
    if isExist {
        db.Delete(task)
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusNotFound)
    }
}

func httpResponse(w http.ResponseWriter, data []byte, statusCode int)  {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    w.Write(data)
}

func DbIsExist(id int) (*Task, bool) {
    var task Task
    err := db.First(&task, id)
    if err.Error == nil {
        return &task, true
    } else {
        return nil, false
    }
}

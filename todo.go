package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
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
)

type TodoServer struct {
    db *gorm.DB
    dbFile string
}

type Task struct {
    ID int          `json:"id"`
    Title string    `json:"title"`
    Completed bool  `json:"completed"`
}

func main() {
    app := TodoServer{}
    router := mux.NewRouter()
    err := app.InitDB()
    app.dbFile = path.Join("database", "todo.db")

    if err != nil {
        panic(err)
    }

    app.RegisterHandlers(router)
    http.ListenAndServe(Port, router)
}

func (app *TodoServer) InitDB() error {
    db, err := gorm.Open(sqlite.Open(app.dbFile), &gorm.Config{})
    if err != nil {
        return err
    }
    app.db = db
    db.AutoMigrate(&Task{})
    return nil
}

func (app *TodoServer) RegisterHandlers(router *mux.Router) {
    router.HandleFunc("/todo", app.GetAllTasks).Methods("GET")
    router.HandleFunc("/todo", app.AddTask).Methods("POST")
    router.HandleFunc("/todo", app.ModifyTask).Methods("PATCH")
    router.HandleFunc("/todo/{id}", app.GetTaskByID).Methods("GET")
    router.HandleFunc("/todo/{id}", app.DeleteTaskByID).Methods("DELETE")
}

func (app *TodoServer) GetAllTasks (w http.ResponseWriter, req *http.Request) {
    tasks := make([]Task, 0)

    err := app.db.Find(&tasks).Error
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    tasksJson, _ := json.MarshalIndent(tasks, "", JsonIndent)
    httpResponse(w, tasksJson, http.StatusOK)
}

func (app *TodoServer) AddTask (w http.ResponseWriter, req *http.Request) {
    var newTask Task
 
    err := json.NewDecoder(req.Body).Decode(&newTask)
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    _, isExist := app.dbIsExist(newTask.ID)
    if isExist {
        http.Error(w, fmt.Sprintf(IDExists, newTask.ID), http.StatusConflict)
        return
    }

    err = app.db.Create(&newTask).Error
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    newTaskJson, _ := json.MarshalIndent(newTask, "", JsonIndent)
    httpResponse(w, newTaskJson, http.StatusCreated)
}

func (app *TodoServer) GetTaskByID (w http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    task, isExist := app.dbIsExist(id)
    if !isExist {
        http.Error(w, fmt.Sprintf(IDNotFound, id), http.StatusNotFound)
        return
    }

    resultJson, _ := json.MarshalIndent(task, "", JsonIndent)
    httpResponse(w, resultJson, http.StatusOK)
}

func (app *TodoServer) ModifyTask (w http.ResponseWriter, req *http.Request) {
    var newTask Task

    err := json.NewDecoder(req.Body).Decode(&newTask)
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    task, isExist := app.dbIsExist(newTask.ID)
    if !isExist {
        http.Error(w, fmt.Sprintf(IDNotFound, newTask.ID), http.StatusNotFound)
        return
    }

    task.Title = newTask.Title
    task.Completed = newTask.Completed
    err = app.db.Save(task).Error
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    resultJson, _ := json.MarshalIndent(task, "", JsonIndent)
    httpResponse(w, resultJson, http.StatusOK)
}

func (app *TodoServer) DeleteTaskByID (w http.ResponseWriter, req *http.Request)  {
    params := mux.Vars(req)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    task, isExist := app.dbIsExist(id)
    if !isExist {
        http.Error(w, fmt.Sprintf(IDNotFound, id), http.StatusNotFound)
    }

    err = app.db.Delete(task).Error
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (app *TodoServer) dbIsExist(id int) (*Task, bool) {
    var task Task
    err := app.db.First(&task, id)
    if err.Error == nil {
        return &task, true
    } else {
        return nil, false
    }
}

func httpResponse(w http.ResponseWriter, data []byte, statusCode int)  {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    w.Write(data)
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)


const (
    JsonIndent = "    "
    Port = ":8080"

    IDExists = "Task with ID %d already exists\n"
    IDNotFound = "Task with ID %d was not found\n"
    BadRequest = "Request is not valid\n"

    DBFile = "DB_FILE.db"
)

type TodoServer struct {
    db *gorm.DB
    dbFilePath string
}

type Task struct {
    ID uint64       `json:"id"`
    Title string    `json:"title"`
    Completed bool  `json:"completed"`
}

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = Port
    } else {
        port = ":" + port
    }

    dbFile := os.Getenv("DB_FILE")
    if dbFile == "" {
        dbFile = DBFile
    }

    app := TodoServer{
    	dbFilePath: DBFile,
    }
    c := cors.AllowAll()
    router := mux.NewRouter()
    err := app.InitDB()

    if err != nil {
        fmt.Fprintf(os.Stderr, err.Error())
        os.Exit(1)
    }

    server := &http.Server{
    	Addr:              Port,
    	Handler:           c.Handler(router),
    }

    router.Use()
    app.RegisterHandlers(router)
    server.ListenAndServe()
}

func (app *TodoServer) InitDB() error {
    db, err := gorm.Open(sqlite.Open(app.dbFilePath), &gorm.Config{})
    if err != nil {
        return err
    }
    app.db = db
    db.AutoMigrate(&Task{})
    return nil
}

func (app *TodoServer) RegisterHandlers(router *mux.Router) {
    router.HandleFunc("/todo", app.GetAllTasks).Methods(http.MethodGet)
    router.HandleFunc("/todo", app.AddTask).Methods(http.MethodPost)
    router.HandleFunc("/todo", app.ModifyTask).Methods(http.MethodPatch)
    router.HandleFunc("/todo/{id}", app.GetTaskByID).Methods(http.MethodGet)
    router.HandleFunc("/todo/{id}", app.DeleteTaskByID).Methods(http.MethodDelete)
}

func (app *TodoServer) GetAllTasks(w http.ResponseWriter, req *http.Request) {
    tasks := make([]Task, 0)

    err := app.db.Find(&tasks).Error
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    tasksJson, _ := json.MarshalIndent(tasks, "", JsonIndent)
    httpResponse(w, tasksJson, http.StatusOK)
}

func (app *TodoServer) AddTask(w http.ResponseWriter, req *http.Request) {
    var newTask Task
 
    err := json.NewDecoder(req.Body).Decode(&newTask)
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    err = validateTaskFields(&newTask)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    newTask.ID = uint64(time.Now().UnixMilli())

    err = app.db.Create(&newTask).Error
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    newTaskJson, _ := json.MarshalIndent(newTask, "", JsonIndent)
    httpResponse(w, newTaskJson, http.StatusCreated)
}

func (app *TodoServer) GetTaskByID(w http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    id, err := strconv.ParseUint(params["id"], 10, 64)
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

func (app *TodoServer) ModifyTask(w http.ResponseWriter, req *http.Request) {
    var newTask Task

    err := json.NewDecoder(req.Body).Decode(&newTask)
    if err != nil {
        http.Error(w, BadRequest, http.StatusBadRequest)
        return
    }

    err = validateTaskFields(&newTask)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
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

func (app *TodoServer) DeleteTaskByID(w http.ResponseWriter, req *http.Request)  {
    params := mux.Vars(req)
    id, err := strconv.ParseUint(params["id"], 10, 64)
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

func (app *TodoServer) dbIsExist(id uint64) (*Task, bool) {
    var task Task
    err := app.db.First(&task, id)
    if err.Error == nil {
        return &task, true
    } else {
        return nil, false
    }
}

func httpResponse(w http.ResponseWriter, data []byte, statusCode int)  {
    w.WriteHeader(statusCode)
    w.Write(data)
}

func validateTaskFields(t *Task) error {
    if len(t.Title) == 0 {
        return errors.New(BadRequest)
    }
    return nil
}

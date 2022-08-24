package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initTestingApp() *TodoServer {
	var err error
    testDatabaseFile := strconv.Itoa(int(time.Now().UnixMicro()))
    app := &TodoServer{
    	dbFilePath: testDatabaseFile,
    }

	if app.db, err = gorm.Open(sqlite.Open(testDatabaseFile), &gorm.Config{}); err != nil {
		fmt.Println("Error")
		return nil
	}
	app.db.AutoMigrate(&Task{})
    return app
}

func removeTestingApp(app *TodoServer) {
	os.RemoveAll(app.dbFilePath)
}

func TestGetAllTasks(t *testing.T) {
    t.Run("Get all todo tasks", func(t *testing.T) {
        var got []Task
        app := initTestingApp()
        defer removeTestingApp(app)

        request := httptest.NewRequest(http.MethodGet, "localhost:8080/todo", nil)
        response := httptest.NewRecorder()

        app.db.Create(&Task{ID: 1, Title: "Clean the room", Completed: false})
        app.GetAllTasks(response, request)

        json.NewDecoder(response.Body).Decode(&got)
        want := []Task{{ID: 1, Title: "Clean the room", Completed: false}}

        assertTaskList(t, got, want)
    })
}

func TestAddTask(t *testing.T) {
    t.Run("Add new task", func(t *testing.T) {
        var got Task
        app := initTestingApp()
        defer removeTestingApp(app)

        reqBody := []byte(`{"id": 1, "title": "Clean the room", "completed": false}`)
        request := httptest.NewRequest(http.MethodPost, "localhost:8080/todo", bytes.NewBuffer(reqBody))
        response := httptest.NewRecorder()
        app.AddTask(response, request)

        json.NewDecoder(response.Body).Decode(&got)
        want := Task{ID: 1, Title: "Clean the room", Completed: false}

        assertTask(t, got, want)
    })

    t.Run("try to add existing element", func(t *testing.T) {
        app := initTestingApp()
        defer removeTestingApp(app)

        postRequestBody := strings.NewReader(`{"id":1, "title":"study", "completed":false}`)
        postRequest := httptest.NewRequest(http.MethodPost, "localhost:8080/todo", postRequestBody)
        postResponse := httptest.NewRecorder()

        task := Task{ID:1, Title :"Clean the room", Completed: false}
        app.db.Create(&task)

        app.AddTask(postResponse, postRequest)
        got := postResponse.Result().Status
        want := "409 Conflict"

        assertString(t, got, want)
    })

    t.Run("Add new task with invalid json syntax", func(t *testing.T) {
        app := initTestingApp()
        defer removeTestingApp(app)

        requestBody := strings.NewReader("invalid json syntax")
		request := httptest.NewRequest(http.MethodPost, "localhost:8080/todo", requestBody)
		response := httptest.NewRecorder()

        app.AddTask(response, request)
        got := response.Result().Status
        want := "400 Bad Request"

        assertString(t, got, want)
    })

    t.Run("Add new task with invalid fields", func(t *testing.T) {
        app := initTestingApp()
        defer removeTestingApp(app)

        requestBody := strings.NewReader(`{"invalid":"invalid", "invalid":"study", "complete":false}`)
        request := httptest.NewRequest(http.MethodPost, "localhost:8080/todo", requestBody)
        response := httptest.NewRecorder()

        app.AddTask(response, request)
        got := response.Result().Status
        want := "400 Bad Request"

        assertString(t, got, want)
    })
}

func TestUpdataTodoItem(t *testing.T) {
	t.Run("Modify existing task", func(t *testing.T) {
		var got []Task
        app := initTestingApp()
        defer removeTestingApp(app)

        requestBody := strings.NewReader(`{"id":1, "title":"Clean the room", "completed":false}`)
		request := httptest.NewRequest(http.MethodPost, "localhost:8080/todo", requestBody)
		response := httptest.NewRecorder()

		patchRequestBody := strings.NewReader(`{"id":1, "title":"Make the bed", "completed":false}`)
		patchRequest := httptest.NewRequest(http.MethodPatch, "localhost:8080/todo", patchRequestBody)
		patchResponse := httptest.NewRecorder()

		getRequest := httptest.NewRequest(http.MethodPost, "localhost:8080/todolist", nil)
		getResponse := httptest.NewRecorder()

		want := []Task{{ID: 1, Title: "Make the bed", Completed: false}}
		app.AddTask(response, request)
		app.ModifyTask(patchResponse, patchRequest)
		app.GetAllTasks(getResponse, getRequest)

		json.NewDecoder(getResponse.Body).Decode(&got)

        assertTaskList(t, got, want)
	})

	t.Run("Modify task with invalid json syntax", func(t *testing.T) {
        app := initTestingApp()
        defer removeTestingApp(app)

		patchRequestBody := strings.NewReader("invalid json syntax")
		patchRequest := httptest.NewRequest(http.MethodPatch, "localhost:8080/todo", patchRequestBody)
		patchResponse := httptest.NewRecorder()

		app.ModifyTask(patchResponse, patchRequest)

		got := patchResponse.Result().Status
		want := "400 Bad Request"

        assertString(t, got, want)
	})

	t.Run("Modify task with invalid fields", func(t *testing.T) {
        app := initTestingApp()
        defer removeTestingApp(app)

		patchRequestBody := strings.NewReader(`{"invalid":"invalid", "invalid":"study", "complete":false}`)
		patchRequest := httptest.NewRequest(http.MethodPatch, "localhost:8080/todo", patchRequestBody)
		patchResponse := httptest.NewRecorder()

		app.ModifyTask(patchResponse, patchRequest)
		got := patchResponse.Result().Status
		want := "400 Bad Request"

        assertString(t, got, want)
	})
}

func TestDeleteTodoItem(t *testing.T) {
	t.Run("Delete existing task", func(t *testing.T) {
        app := initTestingApp()
        defer removeTestingApp(app)

		postRequestBody := strings.NewReader(`{"id":1, "todoItem":"study", "complete":false}`)
		postRequest := httptest.NewRequest(http.MethodPost, "localhost:8080/todo", postRequestBody)
		postResponse := httptest.NewRecorder()

		deleteRequestBody := strings.NewReader(`{"id":1, "todoItem":"study", "complete":false}`)
		deleteRequest := httptest.NewRequest(http.MethodDelete, "localhost:8080/todo/1", deleteRequestBody)
		deleteResponse := httptest.NewRecorder()

		getRequest := httptest.NewRequest(http.MethodPost, "localhost:8080/todo", nil)
		getResponse := httptest.NewRecorder()

		want := []Task{}

		app.AddTask(postResponse, postRequest)
		app.DeleteTaskByID(deleteResponse, deleteRequest)
		app.GetAllTasks(getResponse, getRequest)

		got := []Task{}
		json.NewDecoder(getResponse.Body).Decode(&got)

        assertTaskList(t, got, want)
	})

	t.Run("Delete with invalid task fields", func(t *testing.T) {
        app := initTestingApp()
        defer removeTestingApp(app)

		deleteRequestBody := strings.NewReader(`{"invalid":"invalid", "invalid":"study", "complete":false}`)
		deleteRequest := httptest.NewRequest(http.MethodPatch, "localhost:8080/todolist", deleteRequestBody)
		deleteResponse := httptest.NewRecorder()

		app.DeleteTaskByID(deleteResponse, deleteRequest)
		got := deleteResponse.Result().Status
		want := "400 Bad Request"

        assertString(t, got, want)
	})
}

func assertTaskList(t *testing.T, got, want []Task) {
    t.Helper()

    gotMap := make(map[Task]int)
    wantMap := make(map[Task]int)

    for _, task := range got {
        gotMap[task]++
    }
    
    for _, task := range want {
        wantMap[task]++
    }

    if !reflect.DeepEqual(gotMap, wantMap) {
		t.Errorf("got:\n%v\nwant:\n%v\n", got, want)
    }
}

func assertTask(t *testing.T, got, want Task) {
    t.Helper()

    if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%v\nwant:\n%v\n", got, want)
    }
}

func assertString(t *testing.T, got, want string) {
    t.Helper()

    if got != want {
		t.Errorf("got:\n%q\nwant:\n%q\n", got, want)
    }
}

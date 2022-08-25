package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var s = TodoServer{}

func initDatabase(str string) {

	var err error

	if s.db, err = gorm.Open(sqlite.Open(str+DBFile), &gorm.Config{}); err != nil {
		fmt.Println("Error")
		return
	}
	s.db.AutoMigrate(&Task{})

}

func removeDatabase(str string) {
	os.RemoveAll(str + DBFile)
}

func TestGetAllTasks(t *testing.T) {

	t.Run("try to retrieve all todo items", func(t *testing.T) {

		initDatabase("1")
		defer removeDatabase("1")

		request := httptest.NewRequest(http.MethodGet, "localhost:8080/todolist", nil)
		response := httptest.NewRecorder()

		want := []Task{}
		s.db.Find(&want)

		s.GetAllTasks(response, request)

		got := []Task{}
		json.NewDecoder(response.Body).Decode(&got)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
		}
	})
}

func TestAddTask(t *testing.T) {
	t.Run("try to get a bad request", func(t *testing.T) {
		initDatabase("3")
		defer removeDatabase("3")

		postRequestBody := strings.NewReader(`"ihjkldshl,asdfklskl[sdasdafjssdji()]"false}`)
		postRequest := httptest.NewRequest(http.MethodPost, "localhost:8080/todolist", postRequestBody)
		postResponse := httptest.NewRecorder()

		s.AddTask(postResponse, postRequest)
		got := postResponse.Result().Status
		want := "400 Bad Request"

		if !reflect.DeepEqual(want, got) {
			t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
		}

	})

	t.Run("try to get a bad request again", func(t *testing.T) {

		initDatabase("3")
		defer removeDatabase("3")

		postRequestBody := strings.NewReader(`{"invalid":"invalid", "invalid":"study", "complete":false}`)
		postRequest := httptest.NewRequest(http.MethodPost, "localhost:8080/todolist", postRequestBody)
		postResponse := httptest.NewRecorder()

		s.AddTask(postResponse, postRequest)
		got := postResponse.Result().Status
		want := "400 Bad Request"

		if !reflect.DeepEqual(want, got) {
			t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
		}
	})
}

func TestModifyTask(t *testing.T) {
	t.Run("try to get a bad request in modification", func(t *testing.T) {
		initDatabase("3")
		defer removeDatabase("3")

		patchRequestBody := strings.NewReader(`"ihjkldshl,asdfklskl[sdasdafjssdji()]"false}`)
		patchRequest := httptest.NewRequest(http.MethodPatch, "localhost:8080/todolist", patchRequestBody)
		patchResponse := httptest.NewRecorder()

		s.ModifyTask(patchResponse, patchRequest)
		got := patchResponse.Result().Status
		want := "400 Bad Request"

		if !reflect.DeepEqual(want, got) {
			t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
		}
	})

	t.Run("try to get a bad request in modification again", func(t *testing.T) {
		initDatabase("3")
		defer removeDatabase("3")

		patchRequestBody := strings.NewReader(`{"invalid":"invalid", "invalid":"study", "complete":false}`)
		patchRequest := httptest.NewRequest(http.MethodPatch, "localhost:8080/todolist", patchRequestBody)
		patchResponse := httptest.NewRecorder()

		s.ModifyTask(patchResponse, patchRequest)
		got := patchResponse.Result().Status
		want := "400 Bad Request"

		if !reflect.DeepEqual(want, got) {
			t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
		}
	})
}

func TestDeleteTaskByID(t *testing.T) {
	t.Run("try to delete a specific todo item", func(t *testing.T) {
		initDatabase("5")
		defer removeDatabase("5")

		postRequestBody := strings.NewReader(`{"id":1, "todoItem":"study", "complete":false}`)
		postRequest := httptest.NewRequest(http.MethodPost, "localhost:8080/todolist", postRequestBody)
		postResponse := httptest.NewRecorder()

		deleteRequestBody := strings.NewReader(`{"id":1, "todoItem":"study", "complete":false}`)
		deleteRequest := httptest.NewRequest(http.MethodDelete, "localhost:8080/todolist/1", deleteRequestBody)
		deleteResponse := httptest.NewRecorder()

		getRequest := httptest.NewRequest(http.MethodPost, "localhost:8080/todolist", nil)
		getResponse := httptest.NewRecorder()

		want := []Task{}

		s.AddTask(postResponse, postRequest)
		s.DeleteTaskByID(deleteResponse, deleteRequest)
		s.GetTaskByID(getResponse, getRequest)

		got := []Task{}
		json.NewDecoder(getResponse.Body).Decode(&got)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
		}
	})

	t.Run("try to get a bad request in modification again", func(t *testing.T) {
		initDatabase("3")
		defer removeDatabase("3")

		deleteRequestBody := strings.NewReader(`{"invalid":"invalid", "invalid":"study", "complete":false}`)
		deleteRequest := httptest.NewRequest(http.MethodPatch, "localhost:8080/todolist", deleteRequestBody)
		deleteResponse := httptest.NewRecorder()

		s.DeleteTaskByID(deleteResponse, deleteRequest)
		got := deleteResponse.Result().Status
		want := "400 Bad Request"

		if !reflect.DeepEqual(want, got) {
			t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
		}
	})
}

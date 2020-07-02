package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Constants denoting the state of a task
const (
	start = 2
	pause = 1
	kill  = 0
)

// Response defines the skeleton of response
type Response struct {
	UUID    string `json:"uuid,omitempty"`
	Err     string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Success bool   `json:"success"`
}

// JSONResponse sends the response to client in the form of json
func JSONResponse(w http.ResponseWriter, payload interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

var counter int = 0

// TaskHandler is used to handle different types of requests made by user regarding task handling
type TaskHandler struct {
	logger  *log.Logger
	wg      *sync.WaitGroup
	workers map[string](chan int)
	states  map[string]int
}

// NewTaskHandler creates a new instance of TaskHandler
func NewTaskHandler(l *log.Logger, wg *sync.WaitGroup) *TaskHandler {
	workers := make(map[string](chan int))
	states := make(map[string]int)
	return &TaskHandler{l, wg, workers, states}
}

// CreateTask handles the process of spawning a new task (implemented using goroutine)
func (t *TaskHandler) CreateTask(rw http.ResponseWriter, r *http.Request) {
	t.logger.Println("Endpoint: create")
	rawUUID := uuid.New()
	uuid := strings.Replace(rawUUID.String(), "-", "", -1)
	t.workers[uuid] = make(chan int, 1)
	t.states[uuid] = start
	go t.task(counter, t.workers[uuid], uuid, 10)
	t.workers[uuid] <- start
	t.logger.Println("Task created. uuid:", uuid)
	counter++
	resp := Response{Success: true, UUID: uuid}
	JSONResponse(rw, resp, http.StatusOK)
}

// PauseTask handles the process of pausing a task
func (t *TaskHandler) PauseTask(rw http.ResponseWriter, r *http.Request) {
	uuid := r.Context().Value(KeyUUID{}).(string)
	if t.states[uuid] == pause {
		resp := Response{Success: true, Message: "Already paused"}
		JSONResponse(rw, resp, http.StatusOK)
		return
	}

	t.logger.Println("Endpoint: pause")
	t.workers[uuid] <- pause
	t.states[uuid] = pause
	resp := Response{Success: true}
	JSONResponse(rw, resp, http.StatusOK)
}

// ResumeTask handles the process of resuming a paused task
func (t *TaskHandler) ResumeTask(rw http.ResponseWriter, r *http.Request) {
	uuid := r.Context().Value(KeyUUID{}).(string)
	if t.states[uuid] == start {
		resp := Response{Success: true, Message: "Already runnning"}
		JSONResponse(rw, resp, http.StatusOK)
		return
	}

	t.logger.Println("Endpoint: resume")
	t.workers[uuid] <- start
	t.states[uuid] = start
	resp := Response{Success: true}
	JSONResponse(rw, resp, http.StatusOK)
}

// DeleteTask handles the process of killing an ongoing task
func (t *TaskHandler) DeleteTask(rw http.ResponseWriter, r *http.Request) {
	uuid := r.Context().Value(KeyUUID{}).(string)
	t.logger.Println("Endpoint: delete")
	t.workers[uuid] <- kill
	t.states[uuid] = kill
	resp := Response{Success: true}
	JSONResponse(rw, resp, http.StatusOK)

}

func (t *TaskHandler) closeRoutine(uuid string) {
	t.wg.Done()
	close(t.workers[uuid])
	delete(t.workers, uuid)
	delete(t.states, uuid)
}

// Task is a dummy representation task being spawned by the api. It only logs a statement after a period of 2 seconds
func (t *TaskHandler) task(id int, ch <-chan int, uuid string, cnt int) {
	state := pause
	t.wg.Add(1)
	defer t.closeRoutine(uuid)

	// This for loop is just a representation of how channels can be intergrated to control long running task
	for i := 0; i < cnt; i++ {

		if len(ch) > 0 {
			state = <-ch
			if state == pause {
				t.logger.Println("uuid:", uuid, "status: paused")
				for state == pause {
					state = <-ch
				}
			}

			if state == kill {
				t.logger.Println("uuid:", uuid, "status: killed")
				t.logger.Println("rollback initiated")
				go t.rollBack(uuid)
				return
			}

			t.logger.Println("uuid:", uuid, "status: running")

		}
		// this is used to ensure concurrency by forcing scheduler to rechedule on another task
		runtime.Gosched()

		// Dummy task, can be replaced with actual task to be performed
		t.logger.Println("id:", id, "value:", i)
		time.Sleep(3 * time.Second)

	}

	t.logger.Println("uuid:", uuid, "status: completed")
}

func (t *TaskHandler) rollBack(uuid string) {
	t.wg.Add(1)
	defer t.wg.Done()

	// Dummy task, can be replaced with actual rollback task to be performed
	time.Sleep(1 * time.Second)
	t.logger.Println("uuid:", uuid, "status: Rollback completed")
}

// KillAllTask sends a kill signal to all the running tasks
func (t *TaskHandler) KillAllTask() {
	for k := range t.workers {
		t.workers[k] <- kill
	}
}

// KeyUUID is used to get the data saved in the context
type KeyUUID struct{}

// MiddlewareCheckTask validates the uuid passed with the endpoint and saves it in the context
func (t *TaskHandler) MiddlewareCheckTask(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		i := vars["id"]

		_, ok := t.workers[i]

		if ok == false {
			err := Response{Success: false, Err: "task for given uuid does not exits"}
			JSONResponse(rw, err, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), KeyUUID{}, i)
		r = r.WithContext(ctx)

		next.ServeHTTP(rw, r)

	})
}

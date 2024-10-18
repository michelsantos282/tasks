package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskService struct {
	DB          *sql.DB
	TaskChannel chan Task
}

func (t *TaskService) AddTask(ts *Task) error {
	query := "INSERT INTO tasks (title, description, status, created_at) VALUES (?, ?, ?, ?)"
	result, err := t.DB.Exec(query, ts.Title, ts.Description, ts.Status, ts.CreatedAt)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return err
	}

	ts.ID = int(id)
	return err
}

func (t *TaskService) UpdateTaskStatus(ts Task) error {
	query := "UPDATE tasks set status = ? where id = ?"
	_, err := t.DB.Exec(query, ts.Status, ts.ID)
	return err
}

func (t *TaskService) ListTasks() ([]Task, error) {
	rows, err := t.DB.Query("SELECT * FROM tasks")
	if err != nil {
		return nil, err
	}
	// Só vai ser executado no final da função
	defer rows.Close()
	var tasks []Task

	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t *TaskService) ProcessTasks() {
	for task := range t.TaskChannel {
		log.Printf("Processing task: %s", task.Title)
		time.Sleep(5 * time.Second)
		task.Status = "completed"
		t.UpdateTaskStatus(task)
		log.Printf("Task %s processed", task.Title)
	}
}

func (t *TaskService) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task.Status = "pending"
	task.CreatedAt = time.Now()

	err = t.AddTask(&task)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.TaskChannel <- task
	w.WriteHeader(http.StatusCreated)
}

func (t *TaskService) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := t.ListTasks()
	if err != nil {
		http.Error(w, "Error listing tasks", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func main() {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	taskService := TaskService{
		DB:          db,
		TaskChannel: make(chan Task),
	}

	go taskService.ProcessTasks()

	http.HandleFunc("POST /tasks", taskService.HandleCreateTask)
	http.HandleFunc("GET /tasks", taskService.HandleListTasks)

	log.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)

	// var task Task
	// task.Title = "Test 1"
	// fmt.Println(task.Title)
	// mudaValorPonteiro(&task)
	// fmt.Println(task.Title)
}

// func mudaValorPonteiro(a *Task) {
// 	a.Title = "Test"
// }

// func ponteiros() {
// 	a := 2

// 	fmt.Println(a)

// 	//B igual a referência da memoria de a
// 	b := &a

// 	//Faz a deferencia para mostrar o valor que esta na memoria
// 	fmt.Println(*b)

// 	//Altera o valor que esta na memoria
// 	*b = 5

// 	fmt.Println(*b)
// 	fmt.Println(a)
// }

// func contador(count int) {
// 	for i := range count {
// 		fmt.Println(i)
// 		time.Sleep(time.Second)
// 	}
// }

// Thread principal
// func main() {

// 	//Cria o canal de inteiro com nome da variavel canal
// 	canal := make(chan int)

// 	//T2
// 	go func() {
// 		for i := range 10 {
// 			canal <- i + 10
// 			time.Sleep(time.Second)
// 		}
// 	}()

// 	for v := range canal {
// 		fmt.Println(v)
// 	}

// }

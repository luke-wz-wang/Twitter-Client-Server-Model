package main

import (
	"encoding/json"
	"log"
	"os"
	"proj1/feed"
	"strconv"
	"sync"

)

// struct for decoding request
type Request struct{
	Command string
	Id int
	Body string
	Timestamp float64
}

// struct for encoding normal response
type Response struct{
	Success bool
	Id int
}

// struct for encoding feed contents
type PostContent struct{
	Body string
	Timestamp float64
}

// struct for encoding response to FEED
type FeedContent struct{
	Id int
	Feed []PostContent
}

// TaskQueue struct
type TaskQueue struct {
	Tasks []Request		// stores requests
	Count int		// record the number of requests in queue
	AllDone bool		// indicate whether there are more requests coming; true if not
	Lock sync.Locker
	Cond *sync.Cond
}

func NewTaskQueue() *TaskQueue {
	queue := &TaskQueue{}
	queue.Tasks = nil
	queue.AllDone = false
	queue.Count = 0
	queue.Lock = &sync.Mutex{}
	queue.Cond = sync.NewCond(queue.Lock)
	return queue
}

// add task to task queue; broadcast and set alldone to true if DONE request is received
func producer(taskqueue *TaskQueue){

	dec := json.NewDecoder(os.Stdin)

	var tmp Request
	for{
		err := dec.Decode(&tmp)
		if err != nil {
			return
		}

		taskqueue.Cond.L.Lock()

		if tmp.Command == "DONE"{
			taskqueue.AllDone = true
			taskqueue.Cond.Broadcast()
			taskqueue.Cond.L.Unlock()
			return
		}

		taskqueue.Tasks = append(taskqueue.Tasks, tmp)
		taskqueue.Count++

		taskqueue.Cond.Signal()
		taskqueue.Cond.L.Unlock()
	}
}

// grab blockSize amount of task to process; grab all if not enough;
func consumer(taskqueue *TaskQueue, blockSize int, f feed.Feed, group *sync.WaitGroup){
	for {
		taskqueue.Cond.L.Lock()

		for taskqueue.empty() {
			if taskqueue.AllDone{
				taskqueue.Cond.Signal()
				taskqueue.Cond.L.Unlock()
				group.Done()
				return
			}
			taskqueue.Cond.Wait()
		}

		var tasksAssigned []Request
		if taskqueue.Count <= blockSize {
			tasksAssigned = taskqueue.Tasks
			taskqueue.Tasks = make([]Request, 0)
			taskqueue.Count = 0
		} else {
			tasksAssigned = taskqueue.Tasks[0:blockSize]
			taskqueue.Tasks = taskqueue.Tasks[blockSize:]
			taskqueue.Count -= blockSize
		}

		taskqueue.Cond.Signal()
		taskqueue.Cond.L.Unlock()

		for _, task := range tasksAssigned{
			handleRequest(task, f)
		}
	}

}

// handle a request, update the feed correspondingly, and pose response
func handleRequest(request Request, f feed.Feed){
	enc := json.NewEncoder(os.Stdout)
	if request.Command == "ADD"{
		f.Add(request.Body, request.Timestamp)
		var response Response
		response.Id = request.Id
		response.Success = true
		if err := enc.Encode(&response); err != nil {
			log.Println(err)
		}
	}else if request.Command == "REMOVE"{
		succ := f.Remove(request.Timestamp)
		var response Response
		response.Id = request.Id
		response.Success = succ
		if err := enc.Encode(&response); err != nil {
			log.Println(err)
		}
	}else if request.Command == "CONTAINS"{
		succ := f.Contains(request.Timestamp)
		var response Response
		response.Id = request.Id
		response.Success = succ
		if err := enc.Encode(&response); err != nil {
			log.Println(err)
		}
	}else if request.Command == "FEED"{
		body, timestamps := f.CurrentPosts()
		var posts []PostContent
		for i:= 0; i < len(body); i++{
			var post PostContent
			post.Body = body[i]
			post.Timestamp = timestamps[i]
			posts = append(posts, post)
		}
		var feed FeedContent
		feed.Id = request.Id
		feed.Feed  = posts
		if err := enc.Encode(&feed); err != nil {
			log.Println(err)
		}
	}
}

// check if task queue is empty; return true if it is empty
func (q *TaskQueue)empty() bool{
	if q.Count == 0{
		return true
	}
	return false
}

// receive all requests sequentially
func sequentialProduce(taskqueue *TaskQueue){
	dec := json.NewDecoder(os.Stdin)
	var tmp Request
	for{
		err := dec.Decode(&tmp)
		if err != nil {
			return
		}
		if tmp.Command == "DONE"{
			taskqueue.AllDone = true
			return
		}
		taskqueue.Tasks = append(taskqueue.Tasks, tmp)
		taskqueue.Count++
	}
}
// handle all requests sequentially
func sequentialConsume(taskqueue *TaskQueue,f feed.Feed){
	for _, task := range taskqueue.Tasks{
		handleRequest(task, f)
	}
}


func main() {

	if len(os.Args[1:]) < 2{	// sequential version
		taskQueue := NewTaskQueue()
		f := feed.NewFeed()
		sequentialProduce(taskQueue)
		sequentialConsume(taskQueue, f)
	} else{		// parallel version
		threads, _ := strconv.Atoi(os.Args[1])
		blockSize, _ := strconv.Atoi(os.Args[2])

		var wg sync.WaitGroup
		taskQueue := NewTaskQueue()
		f := feed.NewFeed()

		for i := 0; i < threads; i++ {
			wg.Add(1)
			go consumer(taskQueue, blockSize, f, &wg)
		}

		producer(taskQueue)

		wg.Wait()
	}
}



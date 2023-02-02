package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v9"
	"log"
	"strconv"
	"time"
)

type JobStatus string

var ctx = context.Background()

const (
	SUCCESS JobStatus = "success"
	FAIL    JobStatus = "fail"
	WAIT    JobStatus = "wait"
)
const DefaultWorker = "default"

type Job struct {
	JobId     string              `json:"job_id"`
	Status    JobStatus           `json:"status"`
	Closure   func(job Job) error `json:"-"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

func (j *Job) Marshal() (string, error) {
	marshal, err := json.Marshal(j)
	if err != nil {
		return "", err
	}

	return string(marshal), nil
}

func (j *Job) UnMarshal(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), &j)
	if err != nil {
		return err
	}
	return nil
}

func newJob(jobId string, closure func(job Job) error) Job {
	return Job{
		JobId:   jobId,
		Closure: closure,
		Status:  WAIT,
	}
}

type Queue interface {
	Enqueue(job Job) error
	Dequeue() (*Job, error)
	Clear()
	Count() int
}

type JobQueue struct {
	queue       []Job
	jobChan     chan Job
	maxJobCount int
}

func NewQueue(maxJobCount int) Queue {
	return &JobQueue{
		queue:       make([]Job, 0),
		jobChan:     make(chan Job, maxJobCount),
		maxJobCount: maxJobCount,
	}
}

func (q *JobQueue) Enqueue(job Job) error {
	if q.maxJobCount > len(q.queue) {
		q.queue = append(q.queue, job)
		q.jobChan <- job
		return nil
	}

	return errors.New("can't enqueue job queue: over queue size")
}

func (q *JobQueue) Dequeue() (*Job, error) {
	if len(q.queue) == 0 {
		job := <-q.jobChan
		return &job, nil
	}

	job := q.queue[0]
	q.queue = q.queue[1:]
	jobChan := <-q.jobChan
	if job.JobId == jobChan.JobId {
		return &jobChan, nil
	}

	return nil, errors.New("can't match job id")
}

func (q *JobQueue) Clear() {
	q.queue = make([]Job, 0)
	q.jobChan = make(chan Job)
}

func (q *JobQueue) Count() int {
	return len(q.queue)
}

type Worker interface {
	GetName() string
	Start()
	Stop()
	AddJob(job Job) error
	IsRunning() bool
	MaxJobCount() int
	JobCount() int
}

type JobWorker struct {
	Name        string
	queue       Queue
	jobChan     chan Job
	quitChan    chan bool
	redis       func() *redis.Client
	maxJobCount int
	isRunning   bool
	beforeJob   func(j Job)
	afterJob    func(j Job, err error)
}

func saveJob(r *redis.Client, key string, job Job) {
	jsonJob, err := job.Marshal()
	if err != nil {
		panic(err)
	}

	err = r.Set(ctx, key, jsonJob, time.Minute).Err()
	if err != nil {
		panic(err)
	}
}

func getJob(r *redis.Client, key string) (*Job, error) {
	val, err := r.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		var convJob *Job
		err = convJob.UnMarshal(val)
		if err != nil {
			return nil, err
		}
		return convJob, nil
	}
}

func NewWorker(name string, redis func() *redis.Client, maxJobCount int) Worker {
	return &JobWorker{
		Name:        name,
		queue:       NewQueue(maxJobCount),
		jobChan:     make(chan Job, maxJobCount),
		quitChan:    make(chan bool),
		redis:       redis,
		maxJobCount: maxJobCount,
	}
}

func (w *JobWorker) GetName() string {
	return w.Name
}

func (w *JobWorker) Start() {
	if w.isRunning {
		log.Printf("%s worker is running", w.Name)
		return
	}

	go func() {
		log.Printf("Start Worker(%s):", w.Name)
		for {
			r := w.redis()
			jobChan, err := w.queue.Dequeue()
			if err != nil {
				log.Print(err)
				continue
			}

			w.jobChan <- *jobChan
			select {
			case job := <-w.jobChan:
				key := fmt.Sprintf("%s.%s", w.Name, job.JobId)
				convJob, err := getJob(r, key)
				if err != nil {
					log.Println(err)
				}

				if convJob != nil {
					if convJob.Status != SUCCESS {
						err = w.queue.Enqueue(job)
						if err != nil {
							log.Println(err)
							continue
						}
						continue
					}
				}

				job.CreatedAt = time.Now()
				log.Printf("worker %s, job %s \n", w.Name, job.JobId)
				err = job.Closure(job)
				if err != nil {
					job.Status = FAIL
				} else {
					job.Status = SUCCESS
				}

				job.UpdatedAt = time.Now()
				jsonJob, err := job.Marshal()
				if err != nil {
					log.Print(err)
				}

				log.Printf("end job: %s", jsonJob)
				saveJob(r, key, job)
			case <-w.quitChan:
				log.Printf("worker %s stopping\n", w.Name)
				return
			}
		}
	}()
}

func (w *JobWorker) Stop() {
	if !w.isRunning {
		log.Printf("%s worker is not running", w.Name)
		return
	}

	go func() {
		w.quitChan <- true
	}()
}

func (w *JobWorker) AddJob(job Job) error {
	err := w.queue.Enqueue(job)
	if err != nil {
		return err
	}
	return nil
}

func (w *JobWorker) IsRunning() bool {
	return w.isRunning
}

func (w *JobWorker) MaxJobCount() int {
	return w.maxJobCount
}

func (w *JobWorker) JobCount() int {
	return w.queue.Count()
}

type Dispatcher interface {
	Dispatch(jobId string, closure func(j Job) error) error
	Run()
	Stop()
	SelectWorker(name string) Dispatcher
	GetWorkers() []Worker
	GetRedis() func() *redis.Client
	AddWorker(option Option)
	RemoveWorker(nam string)
	Status(isConsole bool) *Status
}

type JobDispatcher struct {
	workers []Worker
	worker  Worker
	Redis   func() *redis.Client
}

type Option struct {
	Name        string
	MaxJobCount int
	BeforeJob   func(j Job)
	AfterJob    func(j Job)
}

type DispatcherOption struct {
	WorkerOptions []Option
	Redis         func() *redis.Client
}

var defaultWorkerOption = []Option{
	{
		Name:        DefaultWorker,
		MaxJobCount: 10,
	},
}

func NewDispatcher(opt DispatcherOption) Dispatcher {
	workers := make([]Worker, 0)

	if len(opt.WorkerOptions) == 0 {
		opt.WorkerOptions = defaultWorkerOption
	}

	for _, o := range opt.WorkerOptions {
		workers = append(workers, NewWorker(o.Name, opt.Redis, o.MaxJobCount))
	}

	return &JobDispatcher{
		workers: workers,
		worker:  nil,
		Redis:   opt.Redis,
	}
}

func (d *JobDispatcher) AddWorker(option Option) {
	d.workers = append(d.workers, NewWorker(option.Name, d.Redis, option.MaxJobCount))
}

func (d *JobDispatcher) RemoveWorker(name string) {
	var rmIndex *int = nil
	for i, worker := range d.workers {
		if worker.GetName() == name {
			rmIndex = &i
		}
	}

	if rmIndex != nil {
		d.workers = append(d.workers[:*rmIndex], d.workers[*rmIndex+1:]...)
	}
}

func (d *JobDispatcher) GetRedis() func() *redis.Client {
	return d.Redis
}

func (d *JobDispatcher) GetWorkers() []Worker {
	return d.workers
}

func (d *JobDispatcher) SelectWorker(name string) Dispatcher {
	if name == "" {
		for _, w := range d.workers {
			if w.GetName() == "default" {
				d.worker = w
			}
		}

	}

	for _, w := range d.workers {
		if w.GetName() == name {
			d.worker = w
		}
	}

	return d
}

func (d *JobDispatcher) Dispatch(jobId string, closure func(j Job) error) error {
	if d.worker == nil {
		for _, w := range d.workers {
			if w.GetName() == DefaultWorker {
				d.worker = w
			}
		}
	}

	err := d.worker.AddJob(newJob(jobId, closure))
	if err != nil {
		return err
	}

	return nil
}

func (d *JobDispatcher) Run() {
	for _, w := range d.workers {
		w.Start()
	}
}

func (d *JobDispatcher) Stop() {
	for _, w := range d.workers {
		w.Stop()
	}
}

type Status struct {
	Workers     []map[string]string `json:"workers"`
	WorkerCount int                 `json:"worker_count"`
}

func (d *JobDispatcher) Status(isConsole bool) *Status {

	workers := make([]map[string]string, 0)
	for _, w := range d.workers {
		workerInfo := map[string]string{
			"name":          w.GetName(),
			"is_running":    strconv.FormatBool(w.IsRunning()),
			"job_count":     strconv.Itoa(w.JobCount()),
			"max_job_count": strconv.Itoa(w.MaxJobCount()),
		}

		workers = append(workers, workerInfo)
	}

	if isConsole {
		for _, w := range workers {
			log.Printf("[worker name]: %s", w["name"])
			log.Printf("[worker is running]: %s", w["is_running"])
			log.Printf("[worker's job count]: %s", w["job_count"])
			log.Printf("[worker's max job count]:  %s", w["max_job_count"])
		}
		return nil
	}

	return &Status{
		Workers:     workers,
		WorkerCount: len(workers),
	}
}

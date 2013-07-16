package workout

import (
	"sync"
	"time"
)

type JobHandler func(*Job) error
type JobCallback func(*Job, error, time.Duration)


type Master struct {
	ReserveTimeout time.Duration
	concurrency    int
	url            string
	tubes          []string
	workers        []*Worker
	callbacks      map[string]JobCallback
	handlers       map[string]JobHandler
	timeouts       map[string]time.Duration
	job            chan *Job
	quit           chan bool
	mg             sync.WaitGroup
	wg             sync.WaitGroup
}

func NewMaster(url string, tubes []string, concurrency int) *Master {
	return &Master{
		url:         url,
		tubes:       tubes,
		concurrency: concurrency,
		callbacks:   make(map[string]JobCallback),
		handlers:    make(map[string]JobHandler),
		timeouts:    make(map[string]time.Duration),
	}
}

func (m *Master) RegisterHandler(name string, hfn JobHandler, cfn JobCallback, to time.Duration) {
	m.handlers[name] = hfn
	m.callbacks[name] = cfn
	m.timeouts[name] = to
	return
}

func (m *Master) Start() (err error) {
	m.mg.Add(1)

	m.job = make(chan *Job, 2)
	m.quit = make(chan bool, m.concurrency)

	m.workers = make([]*Worker, m.concurrency)

	log.Info("master: starting %d workers...", m.concurrency)

	for i := 0; i < m.concurrency; i++ {
		m.workers[i] = NewWorker(m, i)
		go m.workers[i].run()
	}

	log.Info("master: ready")

	return
}

func (m *Master) Stop() (err error) {
	log.Info("master: stopping %d workers...", m.concurrency)

	for i := 0; i < m.concurrency; i++ {
		m.quit <- true
	}

	m.wg.Wait()
	log.Info("master: %d workers stopped", m.concurrency)

	m.mg.Done()
	log.Info("master: stopped")

	return
}

func (m *Master) Wait() {
	m.mg.Wait()
}

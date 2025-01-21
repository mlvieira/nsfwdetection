package worker

import (
	"runtime"
	"sync"
	"time"

	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/tfmodel"
)

type Job struct {
	ID          int
	FilePath    string
	ResultsChan chan *tfmodel.Prediction
}

var (
	MaxWorkers = runtime.NumCPU()
	MaxJobs    = MaxWorkers * 3
	MaxRetries = 3

	jobQueue    chan Job
	wg          sync.WaitGroup
	once        sync.Once
	shutdown    bool
	shutdownMux sync.Mutex
)

// InitWorkerPool initializes the global job queue and spawns worker goroutines
func InitWorkerPool(nsfwModel *tfmodel.Model) {
	once.Do(func() {
		jobQueue = make(chan Job, MaxJobs)

		for i := 0; i < MaxWorkers; i++ {
			wg.Add(1)
			go workerLoop(i, nsfwModel)
		}
	})
}

// workerLoop continuously processes jobs from the jobQueue.
func workerLoop(workerID int, nsfwModel *tfmodel.Model) {
	defer wg.Done()

	for job := range jobQueue {
		startTime := time.Now()

		prediction, err := processJobWithRetries(workerID, job, nsfwModel)
		duration := float64(time.Since(startTime).Seconds())

		if err != nil {
			sendFailedPrediction(workerID, job, err, duration)
			continue
		}

		sendSuccessfulPrediction(workerID, job, prediction)
	}
}

// processJobWithRetries tries nsfwModel.DetectNSFW up to MaxRetries times.
func processJobWithRetries(workerID int, job Job, nsfwModel *tfmodel.Model) (*tfmodel.Prediction, error) {
	var prediction *tfmodel.Prediction
	var err error

	for retry := 0; retry < MaxRetries; retry++ {
		prediction, err = nsfwModel.DetectNSFW(job.FilePath)
		if err == nil {
			break
		}

		logger.Error("Worker %d: Retry %d failed for job %d: %v",
			workerID, retry+1, job.ID, err)
		time.Sleep(1 * time.Second)
	}
	return prediction, err
}

// sendFailedPrediction logs the critical error and attempts to send an error result on job.ResultsChan.
func sendFailedPrediction(workerID int, job Job, err error, duration float64) {
	logger.Error("Worker %d: Critical error for job %d: %v", workerID, job.ID, err)

	select {
	case job.ResultsChan <- &tfmodel.Prediction{
		ID:        job.ID,
		Error:     err.Error(),
		Trace:     "workerLoop",
		Timestamp: time.Now().Unix(),
		Duration:  duration,
		Success:   false,
	}:
	default:
		logger.Error("Worker %d: Failed to send error result for job %d - channel closed",
			workerID, job.ID)
	}
}

// sendSuccessfulPrediction logs a success and attempts to send a valid prediction on job.ResultsChan.
func sendSuccessfulPrediction(workerID int, job Job, prediction *tfmodel.Prediction) {
	select {
	case job.ResultsChan <- prediction:
	default:
		logger.Error("Worker %d: Failed to send result for job %d - channel closed",
			workerID, job.ID)
	}
}

// SubmitJob places a job on the global queue
func SubmitJob(job Job) {
	shutdownMux.Lock()
	defer shutdownMux.Unlock()

	if shutdown {
		logger.Error("Cannot submit job %d: Worker pool is shutting down", job.ID)
		return
	}
	select {
	case jobQueue <- job:
	default:
		logger.Error("Job queue is full. Dropping job %d", job.ID)
	}
}

// ShutdownWorkerPool closes the global job queue and waits for all workers
func ShutdownWorkerPool() {
	shutdownMux.Lock()
	defer shutdownMux.Unlock()

	if shutdown {
		logger.Info("Worker pool is already shutting down")
		return
	}

	shutdown = true
	close(jobQueue)
	wg.Wait()
	logger.Info("Worker pool shut down successfully")
}

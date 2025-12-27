package controllers

import (
	"net/http"
	"strconv"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// JobsController handles job management endpoints.
type JobsController struct {
	logger    *zap.Logger
	inspector *asynq.Inspector
}

// NewJobsController creates a new JobsController.
func NewJobsController(logger *zap.Logger, redisURL string) (*JobsController, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, err
	}

	inspector := asynq.NewInspector(opt)

	return &JobsController{
		logger:    logger,
		inspector: inspector,
	}, nil
}

// JobInfo represents information about a job.
type JobInfo struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Queue         string                 `json:"queue"`
	State         string                 `json:"state"`
	MaxRetry      int                    `json:"max_retry"`
	Retried       int                    `json:"retried"`
	LastError     string                 `json:"last_error,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	CompletedAt   string                 `json:"completed_at,omitempty"`
	NextProcessAt string                 `json:"next_process_at,omitempty"`
}

// QueueStats represents statistics for a queue.
type QueueStats struct {
	Queue       string `json:"queue"`
	Active      int    `json:"active"`
	Pending     int    `json:"pending"`
	Scheduled   int    `json:"scheduled"`
	Retry       int    `json:"retry"`
	Archived    int    `json:"archived"`
	Completed   int    `json:"completed"`
	Aggregating int    `json:"aggregating"`
	Size        int    `json:"size"`
}

// ListQueues godoc
// @Summary      List all job queues
// @Description  Get statistics for all job queues
// @Tags         Jobs
// @Produce      json
// @Success      200  {array}   QueueStats
// @Failure      500  {object}  map[string]string
// @Router       /jobs/queues [get]
func (jc *JobsController) ListQueues(c echo.Context) error {
	queues, err := jc.inspector.Queues()
	if err != nil {
		jc.logger.Error("Failed to list queues", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list queues"})
	}

	var stats []QueueStats
	for _, queue := range queues {
		info, err := jc.inspector.GetQueueInfo(queue)
		if err != nil {
			jc.logger.Error("Failed to get queue info", zap.String("queue", queue), zap.Error(err))
			continue
		}

		stats = append(stats, QueueStats{
			Queue:       queue,
			Active:      info.Active,
			Pending:     info.Pending,
			Scheduled:   info.Scheduled,
			Retry:       info.Retry,
			Archived:    info.Archived,
			Completed:   info.Completed,
			Aggregating: info.Aggregating,
			Size:        info.Size,
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// ListPendingJobs godoc
// @Summary      List pending jobs
// @Description  Get all pending jobs in a queue
// @Tags         Jobs
// @Produce      json
// @Param        queue  query     string  false  "Queue name"  default(default)
// @Param        limit  query     int     false  "Limit"       default(50)
// @Success      200    {array}   JobInfo
// @Failure      500    {object}  map[string]string
// @Router       /jobs/pending [get]
func (jc *JobsController) ListPendingJobs(c echo.Context) error {
	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	limit := 50
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	tasks, err := jc.inspector.ListPendingTasks(queue, asynq.PageSize(limit))
	if err != nil {
		jc.logger.Error("Failed to list pending tasks", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list pending tasks"})
	}

	var jobs []JobInfo
	for _, task := range tasks {
		jobs = append(jobs, JobInfo{
			ID:       task.ID,
			Type:     task.Type,
			Queue:    task.Queue,
			State:    "pending",
			MaxRetry: task.MaxRetry,
			Retried:  task.Retried,
		})
	}

	return c.JSON(http.StatusOK, jobs)
}

// ListActiveJobs godoc
// @Summary      List active jobs
// @Description  Get all active (running) jobs in a queue
// @Tags         Jobs
// @Produce      json
// @Param        queue  query     string  false  "Queue name"  default(default)
// @Param        limit  query     int     false  "Limit"       default(50)
// @Success      200    {array}   JobInfo
// @Failure      500    {object}  map[string]string
// @Router       /jobs/active [get]
func (jc *JobsController) ListActiveJobs(c echo.Context) error {
	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	limit := 50
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	tasks, err := jc.inspector.ListActiveTasks(queue, asynq.PageSize(limit))
	if err != nil {
		jc.logger.Error("Failed to list active tasks", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list active tasks"})
	}

	var jobs []JobInfo
	for _, task := range tasks {
		jobs = append(jobs, JobInfo{
			ID:       task.ID,
			Type:     task.Type,
			Queue:    task.Queue,
			State:    "active",
			MaxRetry: task.MaxRetry,
			Retried:  task.Retried,
		})
	}

	return c.JSON(http.StatusOK, jobs)
}

// ListScheduledJobs godoc
// @Summary      List scheduled jobs
// @Description  Get all scheduled (future) jobs in a queue
// @Tags         Jobs
// @Produce      json
// @Param        queue  query     string  false  "Queue name"  default(default)
// @Param        limit  query     int     false  "Limit"       default(50)
// @Success      200    {array}   JobInfo
// @Failure      500    {object}  map[string]string
// @Router       /jobs/scheduled [get]
func (jc *JobsController) ListScheduledJobs(c echo.Context) error {
	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	limit := 50
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	tasks, err := jc.inspector.ListScheduledTasks(queue, asynq.PageSize(limit))
	if err != nil {
		jc.logger.Error("Failed to list scheduled tasks", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list scheduled tasks"})
	}

	var jobs []JobInfo
	for _, task := range tasks {
		jobs = append(jobs, JobInfo{
			ID:            task.ID,
			Type:          task.Type,
			Queue:         task.Queue,
			State:         "scheduled",
			MaxRetry:      task.MaxRetry,
			Retried:       task.Retried,
			NextProcessAt: task.NextProcessAt.String(),
		})
	}

	return c.JSON(http.StatusOK, jobs)
}

// ListRetryJobs godoc
// @Summary      List retry jobs
// @Description  Get all jobs pending retry in a queue
// @Tags         Jobs
// @Produce      json
// @Param        queue  query     string  false  "Queue name"  default(default)
// @Param        limit  query     int     false  "Limit"       default(50)
// @Success      200    {array}   JobInfo
// @Failure      500    {object}  map[string]string
// @Router       /jobs/retry [get]
func (jc *JobsController) ListRetryJobs(c echo.Context) error {
	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	limit := 50
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	tasks, err := jc.inspector.ListRetryTasks(queue, asynq.PageSize(limit))
	if err != nil {
		jc.logger.Error("Failed to list retry tasks", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list retry tasks"})
	}

	var jobs []JobInfo
	for _, task := range tasks {
		jobs = append(jobs, JobInfo{
			ID:            task.ID,
			Type:          task.Type,
			Queue:         task.Queue,
			State:         "retry",
			MaxRetry:      task.MaxRetry,
			Retried:       task.Retried,
			LastError:     task.LastErr,
			NextProcessAt: task.NextProcessAt.String(),
		})
	}

	return c.JSON(http.StatusOK, jobs)
}

// ListArchivedJobs godoc
// @Summary      List archived (failed) jobs
// @Description  Get all archived jobs in a queue
// @Tags         Jobs
// @Produce      json
// @Param        queue  query     string  false  "Queue name"  default(default)
// @Param        limit  query     int     false  "Limit"       default(50)
// @Success      200    {array}   JobInfo
// @Failure      500    {object}  map[string]string
// @Router       /jobs/archived [get]
func (jc *JobsController) ListArchivedJobs(c echo.Context) error {
	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	limit := 50
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	tasks, err := jc.inspector.ListArchivedTasks(queue, asynq.PageSize(limit))
	if err != nil {
		jc.logger.Error("Failed to list archived tasks", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list archived tasks"})
	}

	var jobs []JobInfo
	for _, task := range tasks {
		jobs = append(jobs, JobInfo{
			ID:        task.ID,
			Type:      task.Type,
			Queue:     task.Queue,
			State:     "archived",
			MaxRetry:  task.MaxRetry,
			Retried:   task.Retried,
			LastError: task.LastErr,
		})
	}

	return c.JSON(http.StatusOK, jobs)
}

// CancelJob godoc
// @Summary      Cancel a job
// @Description  Cancel a pending or scheduled job
// @Tags         Jobs
// @Produce      json
// @Param        id     path      string  true  "Job ID"
// @Param        queue  query     string  false "Queue name"  default(default)
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /jobs/{id}/cancel [post]
func (jc *JobsController) CancelJob(c echo.Context) error {
	jobID := c.Param("id")
	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	err := jc.inspector.DeleteTask(queue, jobID)
	if err != nil {
		jc.logger.Error("Failed to cancel job",
			zap.String("jobID", jobID),
			zap.String("queue", queue),
			zap.Error(err),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to cancel job"})
	}

	jc.logger.Info("Job cancelled",
		zap.String("jobID", jobID),
		zap.String("queue", queue),
	)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Job cancelled successfully",
		"job_id":  jobID,
	})
}

// RetryJob godoc
// @Summary      Retry an archived job
// @Description  Retry an archived (failed) job immediately
// @Tags         Jobs
// @Produce      json
// @Param        id     path      string  true  "Job ID"
// @Param        queue  query     string  false "Queue name"  default(default)
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /jobs/{id}/retry [post]
func (jc *JobsController) RetryJob(c echo.Context) error {
	jobID := c.Param("id")
	queue := c.QueryParam("queue")
	if queue == "" {
		queue = "default"
	}

	err := jc.inspector.RunTask(queue, jobID)
	if err != nil {
		jc.logger.Error("Failed to retry job",
			zap.String("jobID", jobID),
			zap.String("queue", queue),
			zap.Error(err),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retry job"})
	}

	jc.logger.Info("Job retried",
		zap.String("jobID", jobID),
		zap.String("queue", queue),
	)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Job queued for retry",
		"job_id":  jobID,
	})
}

// PauseQueue godoc
// @Summary      Pause a queue
// @Description  Pause processing of jobs in a queue
// @Tags         Jobs
// @Produce      json
// @Param        queue  path      string  true  "Queue name"
// @Success      200    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /jobs/queues/{queue}/pause [post]
func (jc *JobsController) PauseQueue(c echo.Context) error {
	queue := c.Param("queue")

	err := jc.inspector.PauseQueue(queue)
	if err != nil {
		jc.logger.Error("Failed to pause queue", zap.String("queue", queue), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to pause queue"})
	}

	jc.logger.Info("Queue paused", zap.String("queue", queue))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Queue paused successfully",
		"queue":   queue,
	})
}

// ResumeQueue godoc
// @Summary      Resume a queue
// @Description  Resume processing of jobs in a paused queue
// @Tags         Jobs
// @Produce      json
// @Param        queue  path      string  true  "Queue name"
// @Success      200    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /jobs/queues/{queue}/resume [post]
func (jc *JobsController) ResumeQueue(c echo.Context) error {
	queue := c.Param("queue")

	err := jc.inspector.UnpauseQueue(queue)
	if err != nil {
		jc.logger.Error("Failed to resume queue", zap.String("queue", queue), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to resume queue"})
	}

	jc.logger.Info("Queue resumed", zap.String("queue", queue))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Queue resumed successfully",
		"queue":   queue,
	})
}

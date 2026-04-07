package jobsworker

import "github.com/Mozlook/fotobudka-backend/internal/repository/jobs"

type Worker struct {
	jobsRepo *jobs.Repository
	limit    int
}

func New(jobsRepo *jobs.Repository, limit int) *Worker {
	return &Worker{
		jobsRepo: jobsRepo,
		limit:    limit,
	}
}

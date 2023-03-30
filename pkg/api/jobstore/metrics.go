package jobstore

import "github.com/prometheus/client_golang/prometheus"

var subsystem = "jobstore"

var (
	jobsCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "jobs_created",
			Help:      "Number of jobs in created state.",
		},
	)

	jobsToBeProcessed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "jobs_to_be_processed",
			Help:      "Number of jobs that need processing.",
		},
	)

	jobsProcessing = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "jobs_processing",
			Help:      "Number of jobs being processed.",
		},
	)

	jobsCompleted = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "jobs_succeeded",
			Help:      "Number of jobs that have succeeded.",
		},
	)

	jobsFailed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "jobs_failed",
			Help:      "Number of jobs that have failed.",
		},
	)
)

func init() {
	prometheus.MustRegister(jobsToBeProcessed, jobsCreated, jobsProcessing, jobsCompleted, jobsFailed)
}

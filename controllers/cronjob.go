package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
	"github.com/robfig/cron/v3"

	cronvpav1alpha1 "cronvpa/api/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	scheduleTimeFormat = "Mon, 2006-01-02 15:04:05 -07:00(MST)"
	logTimeFormat      = "2006-01-02 15:04:05 -07:00"
)

var (
	c = cron.New()
)

type CronJob struct {
	JobID         cron.EntryID
	Name          string
	Schedule      string
	ExcludedDates []string
	TimeZone      string
	Suspend       bool
	VPASpec       cronvpav1alpha1.VPASpec
	// cron.Cron
}

func (r *CronVerticalPodAutoscalerReconciler) ConvertToCronJobList(cr *cronvpav1alpha1.CronVerticalPodAutoscaler) []CronJob {
	var (
		cronJobList []CronJob
		timeZone    string
	)
	for _, job := range cr.Spec.Jobs {
		// Setup Time Zone
		if job.TimeZone == "" {
			// Use local time zone
			timeZone = ""
		} else {
			timeZone = strings.Join([]string{"CRON_TZ=", job.TimeZone, " "}, "")
		}
		for _, schedule := range job.Schedules {
			// Setup cronJob
			scheduleWithTimeZone := strings.Join([]string{timeZone, schedule}, "")
			cronJob := CronJob{
				// JobID:         cron.EntryID(1),
				Name:          job.Name,
				Schedule:      scheduleWithTimeZone,
				ExcludedDates: job.ExcludedDates,
				TimeZone:      job.TimeZone,
				Suspend:       job.Suspend,
				VPASpec:       job.VPASpec,
			}
			cronJobList = append(cronJobList, cronJob)
		}
	}
	return cronJobList
}

// Define struct VPAJob for running cron to modify VPA
type VPAJob struct {
	// Ctx context.Context
	context.Context
	Req ctrl.Request
	R   *CronVerticalPodAutoscalerReconciler
	CronJob
	// C       *cron.Cron
}

// Implement cron.Job Interface
func (j VPAJob) Run() {
	location, _ := time.LoadLocation(j.TimeZone)

	// rLog := log.FromContext(j.Context).WithName("vpajob")
	// Apend these values to logs
	rLog := log.FromContext(j.Context).WithValues("Job ID", j.JobID, "Job Name", j.Name, "Schedule", j.Schedule, "Local Time", string(time.Now().In(location).Format(logTimeFormat)))

	// Check if job is suspended.
	if j.Suspend {
		rLog.Info("Suspended job, ignore this schedule")
		return
	}

	rLog.Info("Running cronjob")
	// Check if it is the ExcludedsDates.
	for _, schedule := range j.ExcludedDates {
		if isScheduleDate(schedule, j.TimeZone) {
			rLog.Info("Today is on excluded date, ignore this schedule", "ExcludedDate", schedule)
			return
		}
	}

	// Get cronVPA
	cronVPA := &cronvpav1alpha1.CronVerticalPodAutoscaler{}
	if err := j.R.Get(j.Context, j.Req.NamespacedName, cronVPA); err != nil {
		if errors.IsNotFound(err) {
			// Delete orphan jobs
			rLog.Info("CronVPA was no longer exist, deleting orphan job")
			c.Remove(j.JobID)
			return
			// return ctrl.Result{}, nil
		}
		rLog.Error(err, "Feching cronVPA error")
		return
	} else {
		// Delete invalid jobs
		if !isExistJob(int(j.JobID), cronVPA) {
			rLog.Info("Invalid job,deleting")
			// Delete orphan jobs
			c.Remove(j.JobID)
		}
	}

	patch := client.MergeFrom(cronVPA.DeepCopy())
	// Create or update/patch VPA
	if err := j.R.createOrPatchVPA(j.Context, j.Req, cronVPA, j.CronJob); err != nil {
		rLog.Error(err, "Unable to create/update VPA")
		for i, condition := range cronVPA.Status.Conditions {
			if condition.JobID == int(j.JobID) {
				cronVPA.Status.Conditions[i].LastScheduleTime = string(time.Now().In(location).Format(scheduleTimeFormat))
				cronVPA.Status.Conditions[i].NextScheduleTime = string(c.Entry(j.JobID).Next.In(location).Format(scheduleTimeFormat))
			}
		}
	} else {
		for i, condition := range cronVPA.Status.Conditions {
			if condition.JobID == int(j.JobID) {
				cronVPA.Status.Conditions[i].LastScheduleTime = string(time.Now().In(location).Format(scheduleTimeFormat)) // time.RFC3339
				cronVPA.Status.Conditions[i].LastSuccessfulTime = cronVPA.Status.Conditions[i].LastScheduleTime
				cronVPA.Status.Conditions[i].NextScheduleTime = string(c.Entry(j.JobID).Next.In(location).Format(scheduleTimeFormat))
			}
		}
	}

	// Update job status
	cronVPA.Status.CurrentJob = j.Name
	rLog.Info("Update job condition")
	// if err := j.R.Status().Update(j.Context, cronVPA); err != nil {
	if err := j.R.Status().Patch(j.Context, cronVPA, patch); err != nil {
		if err != nil {
			rLog.Error(err, "Unable to update job condition")
			return
			// return ctrl.Result{}, err
		}
	}
}

func isScheduleDate(schedule string, timeZone string) bool {
	now := time.Now()
	if timeZone == "" {
		timeZone = "Local"
	}
	location, _ := time.LoadLocation(timeZone)

	expr, _ := cronexpr.Parse(schedule)
	scheduleTime := expr.Next(now.In(location))
	return isSameDate(now, scheduleTime)
}

func isSameDate(t1 time.Time, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func isExistJob(jobID int, cr *cronvpav1alpha1.CronVerticalPodAutoscaler) bool {
	for _, condition := range cr.Status.Conditions {
		if jobID == condition.JobID {
			return true
		}
	}
	return false
}

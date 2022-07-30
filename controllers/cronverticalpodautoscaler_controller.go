/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/imdario/mergo"
	"github.com/robfig/cron/v3"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronvpav1alpha1 "cronvpa/api/v1alpha1"
)

// CronVerticalPodAutoscalerReconciler reconciles a CronVerticalPodAutoscaler object
type CronVerticalPodAutoscalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	cronVPAFinalizer     = "cronvpa/finalizer"
	reconcileLoops   int = 0
)

//+kubebuilder:rbac:groups=autoscaling.zoublog.com,resources=cronverticalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling.zoublog.com,resources=cronverticalpodautoscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=autoscaling.zoublog.com,resources=cronverticalpodautoscalers/finalizers,verbs=update
//+kubebuilder:rbac:groups="autoscaling.k8s.io",resources=verticalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="autoscaling.k8s.io",resources=verticalpodautoscalers/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronVerticalPodAutoscaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *CronVerticalPodAutoscalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	// rLog := log.FromContext(ctx).WithValues()
	// rLog := log.FromContext(ctx).WithName("cronvpa")
	rLog := log.FromContext(ctx)
	// Debug level logs, rLog.V(1)

	reconcileLoops = reconcileLoops + 1
	// rLog.Info("+++++++++++Reconcile loops+++++++++++++", "Loops", reconcileLoops)
	rLog.V(0).Info("+++++++++++Reconcile loops+++++++++++++", "Loops", reconcileLoops)

	// var cronVPA *cronvpav1alpha1.CronVerticalPodAutoscaler
	cronVPA := &cronvpav1alpha1.CronVerticalPodAutoscaler{}

	// Fetch cronVPA
	if err := r.Get(ctx, req.NamespacedName, cronVPA); err != nil {
		if errors.IsNotFound(err) {
			rLog.Info("CronVPA deleted")
			return ctrl.Result{}, nil
			// return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, err
	}
	cronVPA_json, _ := json.Marshal(cronVPA.Spec)
	rLog.V(1).Info("Fetched CronVPA", "CronVPA Spec", string(cronVPA_json))

	// Handle the finalizer
	if cronVPA.ObjectMeta.DeletionTimestamp.IsZero() {
		// Register finalizer if it is zero/empty
		if !controllerutil.ContainsFinalizer(cronVPA, cronVPAFinalizer) {
			controllerutil.AddFinalizer(cronVPA, cronVPAFinalizer)
			err := r.Update(ctx, cronVPA)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// If is being deleted
		if controllerutil.ContainsFinalizer(cronVPA, cronVPAFinalizer) {
			// Remove related VPA
			//if err := r.deleteVPAWithCR(ctx, cronVPA); err != nil {
			if err := r.deleteVPAWithReq(ctx, req); err != nil {
				return ctrl.Result{}, err
			}
			rLog.Info("VPA deleted")
			// Deleted jobs
			for _, condition := range cronVPA.Status.Conditions {
				rLog.V(1).Info("Remove job", "Job ID", condition.JobID)
				c.Remove(cron.EntryID(condition.JobID))
			}
			// Remove finalizer
			controllerutil.RemoveFinalizer(cronVPA, cronVPAFinalizer)
			if err := r.Update(ctx, cronVPA); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// CronJobs
	previousStatus := cronVPA.Status.DeepCopy()
	rLog.V(1).Info("Checking status", "Previous Status", previousStatus)

	desiredStatus := cronvpav1alpha1.CronVerticalPodAutoscalerStatus{}
	desiredJobs := r.ConvertToCronJobList(cronVPA)
	rLog.V(1).Info("Desired cronjobs", "CronJob List", desiredJobs)
	// if Job list is not empty, check jobs
	for _, cronJob := range desiredJobs {
		// Assume it is a new job,check if jobs were changed
		newCronJob := true
		for i, condition := range previousStatus.Conditions {
			// Compare Job: Spec is equivalent to condition and Job exist
			if r.CheckJobCondition(cronJob, condition) && r.CheckJobExist(c, condition) {
				rLog.V(1).Info("Existed cronjob", "Condition", condition)
				newCronJob = false
				// If job is not changed, save it to desiredStatus.Conditions
				desiredStatus.Conditions = append(desiredStatus.Conditions, condition)
				// Delete it from previousStatus.Conditions
				previousStatus.Conditions = append(previousStatus.Conditions[:i], previousStatus.Conditions[i+1:]...)
				break
			}
		}
		if newCronJob {
			rLog.V(1).Info("New CronJob", "CronJob Condition", desiredStatus)
			vpaJob := &VPAJob{
				// Ctx:     ctx,
				Context: ctx,
				Req:     req,
				R:       r,
				CronJob: cronJob,
			}
			if jobID, err := c.AddJob(cronJob.Schedule, vpaJob); err != nil {
				rLog.Error(err, "Unable to add job", "Job Name", vpaJob.Name, "Job ID", vpaJob.JobID, "Schedule", vpaJob.Schedule)
				// return ctrl.Result{}, err
			} else {
				c.Start()
				vpaJob.JobID = jobID
				condition := r.generateStatus(int(jobID), cronJob)
				desiredStatus.Conditions = append(desiredStatus.Conditions, condition)
				rLog.Info("Found new/changed, create new job", "Job Name", vpaJob.Name, "Job ID", vpaJob.JobID, "Schedule", vpaJob.Schedule)
			}
		}
	}

	// Jobs still in previousStatus will be deleted
	// Delete old jobs
	for _, condition := range previousStatus.Conditions {
		rLog.Info("Remove old job", "Job ID", condition.JobID)
		c.Remove(cron.EntryID(condition.JobID))
	}

	// jobEntries, _ := json.Marshal(c.Entries())
	// rLog.V(0).Info("All cronjobs", "Entries", string(jobEntries))
	var allJobIDList []cron.EntryID
	var relatedJobIDList []cron.EntryID
	// List cronjob IDs, (bugs) in case some cronjobs are invalid:
	for _, jobEntry := range c.Entries() {
		allJobIDList = append(allJobIDList, jobEntry.ID)
		// invalid := true
		for _, condition := range desiredStatus.Conditions {
			if int(jobEntry.ID) == condition.JobID {
				// invalid = false
				relatedJobIDList = append(relatedJobIDList, jobEntry.ID)
				break
			}
		}
		// if invalid {
		// 	c.Remove(jobEntry.ID)
		// 	rLog.Info("Remove invalid job", "Entry", jobEntry)
		// }
	}
	rLog.V(0).Info("All cronjobs", "EntryIDs", allJobIDList)
	rLog.V(0).Info("Related cronjobs", "EntryIDs", relatedJobIDList)

	// Update Jobs conditions
	if !reflect.DeepEqual(cronVPA.Status.Conditions, desiredStatus.Conditions) {
		rLog.V(1).Info("Update Status", "Desired Status", desiredStatus)
		cronVPA.Status.Conditions = desiredStatus.Conditions
		if err := r.Status().Update(ctx, cronVPA); err != nil {
			return ctrl.Result{}, err
		}
	}

	// VPA
	if cronVPA.Spec.KeepUpdated {
		if err := r.createOrPatchVPA(ctx, req, cronVPA); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronVerticalPodAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cronvpav1alpha1.CronVerticalPodAutoscaler{}).
		Owns(&vpav1.VerticalPodAutoscaler{}).
		Complete(r)
}

// Setup VPA from CR cronVPA
func (r *CronVerticalPodAutoscalerReconciler) setupVPA(cr *cronvpav1alpha1.CronVerticalPodAutoscaler) *vpav1.VerticalPodAutoscaler {
	var (
		jobName string
	)
	if cr.Status.CurrentJob != "" {
		// Using currentJob
		jobName = cr.Status.CurrentJob
	} else {
		jobName = cr.Spec.InitialJob
	}
	// fmt.Println("Job Name of Patch: ", jobName)

	var patchVPA cronvpav1alpha1.VPASpec
	// If `initialJob` not present, will use a empty spec for patching, means only use baseVPA
	for _, job := range cr.Spec.Jobs {
		if job.Name == jobName {
			patchVPA = job.VPASpec
			break
		}
	}

	// Merge baseVPA and patchVPA to mergedVPA
	mergedVPA := cr.Spec.VPASpec
	// fmt.Println("mergedVPA", mergedVPA, patchVPA), mergo.WithOverwriteWithEmptyValue
	mergo.Map(&mergedVPA, patchVPA, mergo.WithOverride)

	vpa := &vpav1.VerticalPodAutoscaler{
		// TypeMeta: metav1.TypeMeta{
		// 	APIVersion: "autoscaling.k8s.io/v1",
		// 	Kind:       "VerticalPodAutoscaler",
		// },
		// ObjectMeta: cr.ObjectMeta,
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cr.Name,
			Labels:    cr.Labels,
		},
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef:      mergedVPA.TargetRef,
			UpdatePolicy:   mergedVPA.PodUpdatePolicy,
			ResourcePolicy: mergedVPA.PodResourcePolicy,
			Recommenders:   mergedVPA.Recommenders,
		},
		// Spec: mergedVPA,
	}

	// https://sdk.operatorframework.io/docs/building-operators/golang/advanced-topics/#handle-cleanup-on-deletion
	// Owner
	// controllerutil.SetControllerReference(cr, vpa, r.Scheme)
	if err := ctrl.SetControllerReference(cr, vpa, r.Scheme); err != nil {
		fmt.Println("SetControllerReference error")
	}

	return vpa
}

func (r *CronVerticalPodAutoscalerReconciler) deleteVPAWithCR(ctx context.Context, cr *cronvpav1alpha1.CronVerticalPodAutoscaler) error {
	// Delete associated VPA
	vpa := r.setupVPA(cr)
	err := r.Delete(ctx, vpa)
	if err != nil {
		// rLog.Error(err, "Unable to delete vpa")
		return err
	}
	// rLog.Info("VPA was deleted")
	return nil
}

func (r *CronVerticalPodAutoscalerReconciler) deleteVPAWithReq(ctx context.Context, req ctrl.Request) error {
	// Delete associated VPA
	vpa := &vpav1.VerticalPodAutoscaler{}
	opts := []client.DeleteAllOfOption{
		client.InNamespace(req.NamespacedName.Namespace),
		client.MatchingFields{"metadata.name": req.NamespacedName.Name},
	}
	err := r.DeleteAllOf(ctx, vpa, opts...)
	if err != nil {
		// rLog.Error(err, "Unable to delete vpa")
		return err
	}
	// rLog.Info("VPA was deleted")
	return nil
}

func (r *CronVerticalPodAutoscalerReconciler) createOrUpdateVPA(ctx context.Context, req ctrl.Request, cr *cronvpav1alpha1.CronVerticalPodAutoscaler, job ...CronJob) error {
	rLog := log.FromContext(ctx)
	if len(job) > 0 {
		location, _ := time.LoadLocation(job[0].TimeZone)
		rLog = rLog.WithValues("Job ID", job[0].JobID, "Job Name", job[0].Name, "Schedule", job[0].Schedule, "Local Time", string(time.Now().In(location).Format(logTimeFormat)))
	}

	desiredVPA := r.setupVPA(cr)
	desiredVPA_json, _ := json.Marshal(desiredVPA)
	rLog.V(1).Info("Creating/Updating VPA", "Desired VPA", string(desiredVPA_json))

	// Create or Update VPA
	currentVPA := &vpav1.VerticalPodAutoscaler{}
	// if err := r.Get(ctx, req.NamespacedName, currentVPA); err != nil {
	if err := r.Get(ctx, types.NamespacedName{Name: desiredVPA.Name, Namespace: desiredVPA.Namespace}, currentVPA); err != nil {
		// If not Found, create desired VPA
		if errors.IsNotFound(err) {
			rLog.Info("VPA not found, Creating VPA")
			if err := r.Create(ctx, desiredVPA); err != nil {
				// if errors.IsAlreadyExists(err) {
				// 	rLog.Error(err, "Already Exists, skip to create VPA")
				// 	return nil
				// }
				rLog.Error(err, "Unable to create VPA")
				return err
			}
			rLog.Info("VPA created")
			return nil
		}
		// If err is not `IsNotFound`, return errors
		return err
	} else {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#update
		// Update VPA
		if !reflect.DeepEqual(desiredVPA.Spec, currentVPA.Spec) {
			currentVPA.Spec = desiredVPA.Spec
			rLog.Info("Updating VPA")
			if err := r.Update(ctx, currentVPA); err != nil {
				// if errors.IsAlreadyExists(err) {
				// 	rLog.Error(err, "\nAlready Exists, Skip to update VPA")
				// 	return nil
				// }
				rLog.Error(err, "Unable to update VPA")
				return err
			}
			rLog.Info("VPA changed")
			rLog.V(1).Info("VPA changed", "VPA Spec", desiredVPA.Spec)
			return nil
		}
		// currentVPA is equal to desiredVPA
		// rLog.Info("VPA do not need to change")
		return nil
	}
}

func (r *CronVerticalPodAutoscalerReconciler) createOrPatchVPA(ctx context.Context, req ctrl.Request, cr *cronvpav1alpha1.CronVerticalPodAutoscaler, job ...CronJob) error {
	rLog := log.FromContext(ctx)
	if len(job) > 0 {
		location, _ := time.LoadLocation(job[0].TimeZone)
		rLog = rLog.WithValues("Job ID", job[0].JobID, "Job Name", job[0].Name, "Schedule", job[0].Schedule, "Local Time", string(time.Now().In(location).Format(logTimeFormat)))
	}

	desiredVPA := r.setupVPA(cr)
	desiredVPA_json, _ := json.Marshal(desiredVPA)
	rLog.V(1).Info("Creating/Patching VPA", "Desired VPA", string(desiredVPA_json))

	// Create or Patch VPA
	currentVPA := &vpav1.VerticalPodAutoscaler{}
	if err := r.Get(ctx, req.NamespacedName, currentVPA); err != nil {
		// If not Found, create desired VPA
		if errors.IsNotFound(err) {
			rLog.Info("VPA not found, Creating VPA")
			if err := r.Create(ctx, desiredVPA); err != nil {
				// if errors.IsAlreadyExists(err) {
				// 	rLog.Error(err, "Already Exists, skip to create VPA")
				// 	return nil
				// }
				rLog.Error(err, "Unable to create VPA")
				return err
			}
			rLog.Info("VPA created")
			return nil
		}
		// If err is not `IsNotFound`, return errors
		return err
	} else {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#patch
		// Patch VPA
		if !reflect.DeepEqual(desiredVPA.Spec, currentVPA.Spec) {
			patch := client.MergeFrom(currentVPA.DeepCopy())
			currentVPA.Spec = desiredVPA.Spec
			rLog.Info("Patching VPA")
			if err := r.Patch(ctx, currentVPA, patch); err != nil {
				// if errors.IsAlreadyExists(err) {
				// 	rLog.Error(err, "\nAlready Exists, Skip to patch VPA")
				// 	return nil
				// }
				rLog.Error(err, "Unable to patch VPA")
				return err
			}
			rLog.Info("VPA changed")
			rLog.V(1).Info("VPA changed", "VPA Spec", desiredVPA.Spec)
			return nil
		}
		// currentVPA is equal to desiredVPA
		// rLog.Info("VPA do not need to change")
		return nil
	}
}

func (r *CronVerticalPodAutoscalerReconciler) generateStatus(jobID int, cronJob CronJob) cronvpav1alpha1.Condtion {
	location, _ := time.LoadLocation(cronJob.TimeZone)
	condition := cronvpav1alpha1.Condtion{
		JobID:              jobID,
		Name:               cronJob.Name,
		Schedule:           cronJob.Schedule,
		ExcludedDates:      cronJob.ExcludedDates,
		TimeZone:           cronJob.TimeZone,
		Suspend:            cronJob.Suspend,
		VPASpec:            cronJob.VPASpec,
		CreationTime:       string(time.Now().In(location).Format(scheduleTimeFormat)),
		NextScheduleTime:   string(c.Entry(cron.EntryID(jobID)).Next.In(location).Format(scheduleTimeFormat)),
		LastScheduleTime:   "",
		LastSuccessfulTime: "",
	}
	return condition
}

func (r *CronVerticalPodAutoscalerReconciler) CheckJobCondition(cronjob CronJob, condition cronvpav1alpha1.Condtion) bool {
	current := CronJob{
		Name:          condition.Name,
		Schedule:      condition.Schedule,
		ExcludedDates: condition.ExcludedDates,
		TimeZone:      condition.TimeZone,
		Suspend:       condition.Suspend,
		VPASpec:       condition.VPASpec,
	}
	if reflect.DeepEqual(cronjob, current) {
		return true
	} else {
		return false
	}
}

func (r *CronVerticalPodAutoscalerReconciler) CheckJobExist(c *cron.Cron, condition cronvpav1alpha1.Condtion) bool {
	return c.Entry(cron.EntryID(condition.JobID)).Valid()
}

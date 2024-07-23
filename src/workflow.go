package osm_download_workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func OsmDownload(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Schedule workflow started.", "StartTime", workflow.Now(ctx))
	so := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: 30 * time.Minute,
	}
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	ao := workflow.ActivityOptions{
		TaskQueue:           "osm-download",
		StartToCloseTimeout: 25 * time.Minute,
		HeartbeatTimeout:    5 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activitySessionCtx := workflow.WithActivityOptions(sessionCtx, ao)

	var a *OsmDownloadActivityObject

	var newPbfActivityResult CheckForNewPbfActivityResult
	var downloadPbfActivityResult DownloadPbfActivityResult
	var uploadPbfActivityResult UploadPbfActivityResult

	err = workflow.ExecuteActivity(activitySessionCtx, a.CheckForNewPbfActivity).Get(activitySessionCtx, &newPbfActivityResult)
	if err != nil {
		return err
	}

	if newPbfActivityResult.NewPbfAvailable {
		err = workflow.ExecuteActivity(activitySessionCtx, a.DownloadPbfActivity).Get(activitySessionCtx, &downloadPbfActivityResult)
		if err != nil {
			return err
		}
		err = workflow.ExecuteActivity(activitySessionCtx, a.UploadPbfActivity, &downloadPbfActivityResult).Get(activitySessionCtx, &uploadPbfActivityResult)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(activitySessionCtx, a.CreateLatestPbfActivity, &uploadPbfActivityResult).Get(activitySessionCtx, nil)
		if err != nil {
			return err
		}

		// cwo := workflow.ChildWorkflowOptions{
		// 	WorkflowExecutionTimeout: 10 * time.Minute,
		// 	TaskQueue:                "osm-extractor",
		// 	WorkflowID:               "extract-osm-cutouts" + "-" + uuid.New().String(),
		// }
		// ctx = workflow.WithChildOptions(ctx, cwo)
		// err := workflow.ExecuteChildWorkflow(ctx, "extract-osm-cutouts").Get(ctx, nil)
		// if err != nil {
		// 	return err
		// }
	}
	return nil
}

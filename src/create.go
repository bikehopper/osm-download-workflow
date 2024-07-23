package osm_download_workflow

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"

	app_config "github.com/bikehopper/osm-download-workflow/src/app_config"
)

func Create() {
	ctx := context.Background()
	conf := app_config.New()
	hostPort := conf.TemporalUrl
	temporalClient, err := client.Dial(client.Options{
		HostPort: hostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal Client", err)
	}
	defer temporalClient.Close()

	// Create Schedule and Workflow IDs
	scheduleID := "osm-download-schedule"
	workflowID := "osm-download"
	catchupDuration, _ := time.ParseDuration("1h")
	jitterDuration, _ := time.ParseDuration("2m")

	spec := client.ScheduleSpec{
		Calendars: []client.ScheduleCalendarSpec{
			{Hour: []client.ScheduleRange{{Start: 0, End: 23}}},
		},
		Jitter:       jitterDuration,
		TimeZoneName: "US/Pacific",
	}
	// Create the schedule
	_, err = temporalClient.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID:            scheduleID,
		CatchupWindow: catchupDuration,
		Overlap:       enums.SCHEDULE_OVERLAP_POLICY_SKIP,
		Spec:          spec,
		Action: &client.ScheduleWorkflowAction{
			ID:        workflowID + "-" + uuid.New().String(),
			Workflow:  OsmDownload,
			TaskQueue: "osm-download",
		},
	})
	if err != nil && err.Error() != "schedule with this ID is already registered" {
		log.Fatalln("Unable to create schedule", err)
	}
	log.Println("Schedule created", "ScheduleID", scheduleID)
}

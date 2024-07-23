package osm_download_workflow

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	app_config "github.com/bikehopper/osm-download-workflow/src/app_config"
)

func Worker() {
	conf := app_config.New()
	hostPort := conf.TemporalUrl
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: hostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "osm-download", worker.Options{
		EnableSessionWorker: true,
	})

	var activities *OsmDownloadActivityObject

	w.RegisterWorkflow(OsmDownload)
	w.RegisterActivity(activities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

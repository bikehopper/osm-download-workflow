package osm_download_workflow

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.temporal.io/sdk/activity"

	app_config "github.com/bikehopper/osm-download-workflow/src/app_config"
)

type OsmDownloadActivityObject struct{}

type CheckForNewPbfActivityResult struct {
	NewPbfAvailable bool
}

type DownloadPbfActivityResult struct {
	FilePath string
	Etag     string
}

type UploadPfbActivityParams struct {
	FilePath string
	Etag     string
}

type UploadPbfActivityResult struct {
	Key string
}

type CreateLatestPbfActivityParams struct {
	Key string
}

type FetchPbfResult struct {
	File *os.File
	Etag string
}

func getEtagHttp(url string) (*string, error) {
	latestHeadRes, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	if latestHeadRes.StatusCode > 299 {
		return nil, err
	}

	latestEtag := latestHeadRes.Header.Get("ETag")
	return &latestEtag, nil
}

func fetchPbf(url string) (*FetchPbfResult, error) {
	file, err := os.CreateTemp("", "us-latest.*.osm.pbf")
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}
	return &FetchPbfResult{
		File: file,
		Etag: resp.Header.Get("ETag"),
	}, nil
}

func getDatedFileName(filePath string, date time.Time) string {
	fileName := filepath.Base(filePath)
	return date.Format("2006-01-02") + "-" + strings.Replace(fileName, "-latest", "", 1)
}

func (o *OsmDownloadActivityObject) CheckForNewPbfActivity(ctx context.Context) (*CheckForNewPbfActivityResult, error) {
	// logger := activity.GetLogger(ctx)
	conf := app_config.New()

	// Get Etag of latest PBF from Geofabrik
	result := &CheckForNewPbfActivityResult{
		NewPbfAvailable: false,
	}
	latestEtag, err := getEtagHttp(conf.PbfUrl)
	if err != nil {
		return nil, err
	}

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = conf.S3Region
		o.BaseEndpoint = aws.String(conf.S3EndpointUrl)
		o.UsePathStyle = true
	})

	lastUpdatedPbfHead, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(conf.Bucket),
		Key:    aws.String(conf.PbfKey),
	})
	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			result.NewPbfAvailable = true
			return result, nil
		}
		return result, err
	}

	lastUpdatedPbfEtag := lastUpdatedPbfHead.Metadata["geofabrik-etag"]
	if *latestEtag != lastUpdatedPbfEtag {
		result.NewPbfAvailable = true
	}

	return result, nil
}

func (o *OsmDownloadActivityObject) DownloadPbfActivity(ctx context.Context) (*DownloadPbfActivityResult, error) {
	conf := app_config.New()

	fetchResult, err := fetchPbf(conf.PbfUrl)
	if err != nil {
		return nil, err
	}
	defer fetchResult.File.Close()
	result := &DownloadPbfActivityResult{
		FilePath: fetchResult.File.Name(),
		Etag:     fetchResult.Etag,
	}

	return result, nil
}

func (o *OsmDownloadActivityObject) UploadPbfActivity(ctx context.Context, param UploadPfbActivityParams) (*UploadPbfActivityResult, error) {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	conf := app_config.New()
	s3Client := s3.NewFromConfig(cfg, func(opt *s3.Options) {
		opt.Region = conf.S3Region
		opt.BaseEndpoint = aws.String(conf.S3EndpointUrl)
		opt.UsePathStyle = true
	})

	datedFileName := getDatedFileName(conf.PbfKey, activity.GetInfo(ctx).ScheduledTime)
	objectKey := strings.Replace(conf.PbfKey, path.Base(conf.PbfKey), datedFileName, 1)
	fileToUpload, err := os.Open(param.FilePath)
	if err != nil {
		return nil, err
	}
	// not really necc. when run in a container but useful otherwise
	defer os.Remove(fileToUpload.Name())
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(conf.Bucket),
		Key:    aws.String(objectKey),
		Body:   fileToUpload,
		Metadata: map[string]string{
			"geofabrik-etag": param.Etag,
		},
	})
	if err != nil {
		return nil, err
	}

	return &UploadPbfActivityResult{
		Key: objectKey,
	}, nil
}

func (o *OsmDownloadActivityObject) CreateLatestPbfActivity(ctx context.Context, params CreateLatestPbfActivityParams) error {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	conf := app_config.New()
	s3Client := s3.NewFromConfig(cfg, func(opt *s3.Options) {
		opt.Region = conf.S3Region
		opt.BaseEndpoint = aws.String(conf.S3EndpointUrl)
		opt.UsePathStyle = true
	})

	_, err := s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		CopySource: aws.String(filepath.Join(conf.Bucket, params.Key)),
		Bucket:     aws.String(conf.Bucket),
		Key:        aws.String(conf.PbfKey),
	})
	if err != nil {
		return err
	}

	return nil
}

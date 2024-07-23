package app_config

import "os"

type AppConfig struct {
	PbfUrl        string
	S3Region      string
	S3EndpointUrl string
	Bucket        string
	PbfKey        string
	TemporalUrl   string
}

func New() *AppConfig {
	return &AppConfig{
		PbfUrl:        getEnv("PBF_URL", ""),
		S3Region:      getEnv("S3_REGION", "us-east-1"),
		S3EndpointUrl: getEnv("S3_ENDPOINT_URL", ""),
		Bucket:        getEnv("BUCKET", ""),
		PbfKey:        getEnv("PBF_KEY", ""),
		TemporalUrl:   getEnv("TEMPORAL_URL", "localhost:7233"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

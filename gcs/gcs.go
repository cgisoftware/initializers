package gcs

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCSConfig struct {
	*storage.Client
}

type GCSClientConfig struct {
	context  context.Context
	filePath string
}

type GCSClientOption func(d *GCSClientConfig)

func Initialize(ctx context.Context, filePath string, opts ...GCSClientOption) *GCSConfig {
	gcsOptions := &GCSClientConfig{}
	gcsOptions.context = ctx
	gcsOptions.filePath = filePath
	for _, opt := range opts {
		opt(gcsOptions)
	}

	client, err := storage.NewClient(gcsOptions.context, option.WithCredentialsFile(gcsOptions.filePath))
	if err != nil {
		panic(err)
	}

	return &GCSConfig{client}
}

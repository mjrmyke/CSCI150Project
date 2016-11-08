package main

import (
	"io"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

func AddFileToGCS(ctx context.Context, filename, contentType string, freader io.Reader) error {
	client, clientErr := storage.NewClient(ctx)
	if clientErr != nil {
		return clientErr
	}
	defer client.Close()

	csWriter := client.Bucket(gcsBucket).Object(filename).NewWriter(ctx)

	// Cloud Storage Writer - Permissions
	csWriter.ACL = []storage.ACLRule{
		{storage.AllUsers, storage.RoleReader},
	}

	csWriter.ContentType = contentType
	if _, copyErr := io.Copy(csWriter, freader); copyErr != nil {
		csWriter.Close()
		return copyErr
	}
	return csWriter.Close()
}

func RemoveFileFromGCS(ctx context.Context, filename string) error {
	client, clientErr := storage.NewClient(ctx)
	if clientErr != nil {
		return clientErr
	}
	defer client.Close()
	return client.Bucket(gcsBucket).Object(filename).Delete(ctx)
}

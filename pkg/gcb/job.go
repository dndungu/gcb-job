package gcb

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudbuild/v1"
	"io/ioutil"
	"time"
)

type Job struct {
	commitSha        string
	projectID        string
	repoName         string
	triggerID        string
	subscriptionName string
	status           chan string
}

func (j Job) Run() (id string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpClient, err := google.DefaultClient(ctx, oauthScope)
	if err != nil {
		return
	}

	rs := cloudbuild.RepoSource{
		CommitSha: j.commitSha,
		ProjectId: j.projectID,
		RepoName:  j.repoName,
		Dir:       "/",
	}
	requestBody, err := rs.MarshalJSON()
	if err != nil {
		return
	}
	reader := bytes.NewReader(requestBody)
	u := fmt.Sprintf(triggerURL, j.projectID, j.triggerID)

	response, err := httpClient.Post(u, "application/json", reader)
	if err != nil {
		return
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	var op cloudbuild.Operation
	err = json.Unmarshal(responseBody, &op)
	if err != nil {
		return
	}

	if response.StatusCode < 199 || response.StatusCode > 299 {
		return "", fmt.Errorf(
			"Received status code %d and body %s after sending %s to %s",
			response.StatusCode,
			responseBody,
			string(requestBody),
			u,
		)
	}

	b, err := op.Metadata.MarshalJSON()
	if err != nil {
		return
	}

	var m cloudbuild.BuildOperationMetadata
	err = json.Unmarshal(b, &m)
	if err == nil {
		id = m.Build.Id
	}
	return
}

func (j Job) Wait(id string) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubsubClient, err := pubsub.NewClient(ctx, j.projectID)
	if err != nil {
		return
	}
	subscription := pubsubClient.Subscription(j.subscriptionName)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		return
	}

	if !exists {
		return fmt.Errorf(
			"could not find subscription '%s' in project %s",
			j.subscriptionName,
			j.projectID,
		)
	}

	errChan := make(chan error, 1)
	go func() {
		subscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
			var build cloudbuild.Build
			if err := json.Unmarshal(m.Data, &build); err != nil {
				errChan <- err
				return
			}

			defer m.Ack()

			if id != build.Id {
				return
			}

			j.status <- build.Status

			switch build.Status {
			case "QUEUED":
			case "WORKING":
			case "SUCCESS":
				errChan <- nil
				close(j.status)
				close(errChan)
			default:
				errChan <- errors.New(build.Status)
				close(j.status)
				close(errChan)
			}
		})
	}()
	return <-errChan
}

func (j Job) Status() chan string {
	return j.status
}

func New(options ...Option) *Job {
	j := Job{status: make(chan string)}
	for _, option := range options {
		option(&j)
	}
	return &j
}

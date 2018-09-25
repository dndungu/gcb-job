package main

import (
	"github.com/dndungu/spinnaker-gcb-stage/pkg/gcb"
	"github.com/zatiti/log"
	"os"
)

var GitRevision string

type Job interface {
	Run() (id string, err error)
	Status() chan string
	Wait(id string) error
}

var job Job

var logger = log.New()

func env(k string) string {
	v, ok := os.LookupEnv(k)
	if !ok {
		logger.Fatalf("%s is not set in environment.", k)
	}
	return v
}

func init() {
	commitSha := env("COMMIT_SHA")
	logger.WithField("commit_sha", commitSha)

	projectID := env("PROJECT_ID")
	logger.WithField("project_id", projectID)

	repoName := env("REPO_NAME")
	logger.WithField("repo_name", repoName)

	triggerID := env("TRIGGER_ID")
	logger.WithField("trigger_id", triggerID)

	subscriptionName := env("SUBSCRIPTION_NAME")
	logger.WithField("subscription_name", subscriptionName)

	job = gcb.New(
		gcb.WithCommitSha(commitSha),
		gcb.WithProjectId(projectID),
		gcb.WithRepoName(repoName),
		gcb.WithTriggerId(triggerID),
		gcb.WithSubscriptionName(subscriptionName),
	)
}

func main() {
	id, err := job.Run()

	done := make(chan bool)
	go func() {
		for status := range job.Status() {
			logger.Info(status)
		}
		done <- true
	}()

	if err == nil {
		err = job.Wait(id)
	}

	if err != nil {
		logger.Fatal(err.Error())
	}

	<-done
}

package gcb

type Option func(*Job)

func WithCommitSha(v string) Option {
	return func(j *Job) {
		j.commitSha = v
	}
}

func WithProjectId(v string) Option {
	return func(j *Job) {
		j.projectID = v
	}
}

func WithRepoName(v string) Option {
	return func(j *Job) {
		j.repoName = v
	}
}

func WithTriggerId(v string) Option {
	return func(j *Job) {
		j.triggerID = v
	}
}

func WithSubscriptionName(v string) Option {
	return func(j *Job) {
		j.subscriptionName = v
	}
}

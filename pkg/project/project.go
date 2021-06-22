package project

var (
	description string = "The azure-collector exposes Azure metrics."
	gitSHA             = "n/a"
	name        string = "azure-collector"
	source      string = "https://github.com/giantswarm/azure-collector"
	version            = "2.5.1"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}

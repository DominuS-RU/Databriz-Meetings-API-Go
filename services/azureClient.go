package services

import (
	"github.com/dghubble/sling"
	"net/http"
)

const (
	azureAPI = "https://dev.azure.com/"
)

type AzureClient struct {
	sling     *sling.Sling
	Projects  *ProjectsService
	Teams     *TeamsService
	WorkItems *WorkItemsService
}

func NewAzureClient(token string, organization string) *AzureClient {
	httpClient := &http.Client{}
	base := sling.New().Client(httpClient).Base(azureAPI).SetBasicAuth("", token).
		Set("Accept", "application/json")

	return &AzureClient{
		sling:     base,
		Projects:  newProjectsService(base.New(), organization),
		Teams:     newTeamsService(base.New(), organization),
		WorkItems: newWorkItemsService(base.New(), organization),
	}
}

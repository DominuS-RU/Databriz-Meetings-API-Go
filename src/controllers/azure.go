package controllers

import (
	"../config"
	"../httputil"
	"../models"
	"../services"
	"../utils"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

var interactor utils.AzureInteractor
var client *services.AzureClient

type AzureController struct{}

func NewAzureController() *AzureController {

	client = services.NewAzureClient(
		viper.GetString(config.AzureToken),
		viper.GetString(config.AzureOrganization),
	)

	interactor = utils.AzureInteractor{
		OrganizationName: viper.GetString(config.AzureOrganization),
		AuthToken:        viper.GetString(config.AzureToken),
	}

	return &AzureController{}
}

func (c *AzureController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("projects/list", c.getProjectsList)

	router.GET("teams/list", c.getProjectTeams)              // ?projectId
	router.GET("teams/members/list", c.getTeamMembers)       // ?projectId, teamId
	router.GET("teams/iterations/list", c.getTeamIterations) // ?projectId, teamId

	router.GET("members/:memberId/workItems/list", c.getMemberWorkItems) // ?projectId, teamId, iteration
}

// @Summary Список проектов
// @Description Возвращает список проектов организации
// @Tags Azure
// @Produce json
// @Success 200 {array} models.Project
// @Failure 500 {object} httputil.HTTPError "When failed to receive data from Azure"
// @Router /v1/azure/projects/list [get]
func (AzureController) getProjectsList(ctx *gin.Context) {
	projects, _, err := client.Projects.Projects(
		&services.ProjectsParams{ApiVersion: "5.1"},
	)
	if err != nil {
		httputil.NewInternalAzureError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, models.FromAzureProjectsList(projects))
}

// @Summary Список команд
// @Description Возвращает список команд проекта
// @Tags Azure
// @Produce json
// @Param projectId query string true "Project Id"
// @Success 200 {array} models.Team
// @Failure 400 {object} httputil.HTTPError "When user has not provided projectId parameter"
// @Failure 500 {object} httputil.HTTPError "When failed to receive data from Azure"
// @Router /v1/azure/teams/list [get]
func (AzureController) getProjectTeams(ctx *gin.Context) {
	projectId := ctx.Query("projectId")

	if projectId == "" {
		httputil.NewError(ctx, http.StatusBadRequest, "projectId must be provided")
		return
	}

	projectTeams, _, err := client.Projects.ProjectTeams(
		projectId,
		&services.ProjectsParams{ApiVersion: "5.0"},
	)

	if err != nil {
		httputil.NewInternalAzureError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, models.FromAzureTeamsList(projectTeams))
}

// @Summary Список участников команды
// @Description Возвращает список участников команды проекта
// @Tags Azure
// @Produce json
// @Param projectId query string true "Project Id"
// @Param teamId query string true "Team Id"
// @Success 200 {array} models.Member
// @Failure 400 {object} httputil.HTTPError "When user has not provided projectId or teamId parameter"
// @Failure 500 {object} httputil.HTTPError "When failed to receive data from Azure"
// @Router /v1/azure/teams/members/list [get]
func (AzureController) getTeamMembers(ctx *gin.Context) {
	projectId := ctx.Query("projectId")
	teamId := ctx.Query("teamId")

	if projectId == "" || teamId == "" {
		httputil.NewError(ctx, http.StatusBadRequest, "projectId and teamId must be provided")
		return
	}

	teamMembers, _, err := client.Teams.TeamMembers(
		projectId,
		teamId,
		&services.TeamsParams{ApiVersion: "5.1"},
	)

	if err != nil {
		httputil.NewInternalAzureError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, models.FromAzureMembersList(teamMembers))
}

// @Summary Список спринтов команды
// @Description Возвращает список спринтов команды
// @Tags Azure
// @Produce json
// @Param projectId query string true "Project Id"
// @Param teamId query string true "Team Id"
// @Success 200 {array} models.Iteration
// @Failure 400 {object} httputil.HTTPError "When user has not provided projectId or teamId parameter"
// @Failure 500 {object} httputil.HTTPError "When failed to receive data from Azure"
// @Router /v1/azure/teams/iterations/list [get]
func (AzureController) getTeamIterations(ctx *gin.Context) {
	projectId := ctx.Query("projectId")
	teamId := ctx.Query("teamId")

	if projectId == "" || teamId == "" {
		httputil.NewError(ctx, http.StatusBadRequest, "projectId and teamId must be provided")
		return
	}

	teamIterations, _, err := client.Teams.TeamIterations(
		projectId,
		teamId,
		&services.TeamsParams{ApiVersion: "5.1"},
	)

	if err != nil {
		httputil.NewInternalAzureError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, models.FromAzureIterations(teamIterations))
}

// @Summary Задачи определенного участника команды
// @Tags Azure
// @Produce json
// @Param userId path string true "User Email"
// @Param projectId query string true "Project Id"
// @Param teamId query string true "Team Id"
// @Param iteration query string true "Iteration Name"
// @Success 200 {object} azure.WorkItemsResponse
// @Failure 400 {object} httputil.HTTPError "When user has not provided projectId or teamId parameter"
// @Failure 500 {object} httputil.HTTPError "When failed to receive data from Azure"
// @Router /v1/azure/members/{memberId}/workItems/list [get]
func (AzureController) getMemberWorkItems(ctx *gin.Context) {
	userEmail := ctx.Param("memberId")
	projectId := ctx.Query("projectId")
	teamId := ctx.Query("teamId")
	iteration := ctx.Query("iteration")

	if projectId == "" || teamId == "" || userEmail == "" || iteration == "" {
		httputil.NewError(ctx, http.StatusBadRequest, "projectId, teamId, userId, iteration must be provided")
		return
	}

	workItemsWiql, err := interactor.GetWorkItemsByWiql(projectId, teamId, userEmail, iteration)
	if err != nil {
		httputil.NewInternalAzureError(ctx)
		return
	}

	// Request detailed works
	var newList = make([]int, len(workItemsWiql.WorkItems))
	for index, element := range workItemsWiql.WorkItems {
		newList[index] = element.ID
	}

	workItems, err := interactor.GetWorkItemsDescription(projectId, newList)
	if err != nil {
		httputil.NewInternalAzureError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, workItems)
}

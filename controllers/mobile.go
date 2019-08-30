package controllers

import (
	"Databriz-Meetings-API-Go/db"
	"Databriz-Meetings-API-Go/httputil"
	"Databriz-Meetings-API-Go/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type MobileController struct{}

func NewMobileController() *MobileController {
	return &MobileController{}
}

func (c *MobileController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("control/show", c.showMemberWorkItems)
}

// @Summary Переключение фронта
// @Description Переключает отображающегося ппользователя на фронте
// @Tags Mobile
// @Produce json
// @Success 200
// @Failure 400 {object} httputil.HTTPError "When user has not provided correct response body"
// @Router /v1/azure/teams/list [get]
func (MobileController) showMemberWorkItems(ctx *gin.Context) {
	var requestBody models.ShowRequestBody
	if err := ctx.BindJSON(&requestBody); err != nil {
		log.Println(err.Error())
		httputil.NewError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	db.StoreData(requestBody)

	ctx.JSON(http.StatusOK, struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	})
}

package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/synera-br/golang-cloud-collector/internal/core/service"
)

type AzureHandlerHttpInterface interface {
	ListResources(c *gin.Context)
	FindByResourceGroup(c *gin.Context)
	FindByTag(c *gin.Context)
	GetSubscription(c *gin.Context)
}

type AzureHandlerHttp struct {
	Service service.AzureServiceInterface
}

func NewAzureHandlerHttp(svc service.AzureServiceInterface, routerGroup *gin.RouterGroup) AzureHandlerHttpInterface {

	azure := &AzureHandlerHttp{
		Service: svc,
	}

	azure.handlers(routerGroup)

	return azure
}

func (c *AzureHandlerHttp) handlers(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/azure", c.ListResources)
	routerGroup.GET("/azure/:name", c.FindByResourceGroup)
	routerGroup.GET("/azure/tags", c.FindByTag)
	routerGroup.GET("/azure/subscription/:name", c.GetSubscription)

}

// AzureListResources    godoc
// @Summary     list all resources from subscription
// @Tags        azure
// @Accept       json
// @Produce     json
// @Description get all azure register
// @Success     200 {object} []interface{}
// @Failure     404 {object} string
// @Failure     500 {object} string
// @Router      /azure [get]
func (obj *AzureHandlerHttp) ListResources(c *gin.Context) {

	result, err := obj.Service.ListResources(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, result)
}

// AzureFindByResourceGroup    godoc
// @Summary     list all resources from resource group
// @Tags        azure
// @Accept       json
// @Produce     json
// @Param       name path string true "name"
// @Description get all azure register
// @Success     200 {object} []interface{}
// @Failure     404 {object} string
// @Failure     500 {object} string
// @Router      /azure/{name} [get]
func (obj *AzureHandlerHttp) FindByResourceGroup(c *gin.Context) {
	rsg := c.Param("name")

	if len(rsg) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "resource group name not setted"})
		return
	}

	result, err := obj.Service.ListResourcesByResourceGroup(context.Background(), rsg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, "not found")
		return
	}

	c.JSON(http.StatusAccepted, result)

}

// AzureFindByTag    godoc
// @Summary     list resources filter by tags
// @Tags        azure
// @Accept       json
// @Produce     json
// @Description find resources by tags
// @Param key        query string false "Key filter"
// @Param value        query string false "value filter"
// @Success     200 {object} []interface{}
// @Failure     404 {object} string
// @Failure     500 {object} string
// @Router      /azure/tags [get]
func (obj *AzureHandlerHttp) FindByTag(c *gin.Context) {

	if len(c.Request.URL.Query()) != 2 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "key and value of tags not setted"})
		return
	}

	result, err := obj.Service.ListResourcesByTag(context.Background(), c.Request.URL.Query().Get("key"), c.Request.URL.Query().Get("value"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, "not found")
		return
	}

	c.JSON(http.StatusAccepted, result)
}

// AzureGetSubscription    godoc
// @Summary     get subscription information
// @Tags        azure
// @Accept       json
// @Produce     json
// @Description get subscription information
// @Param       name path string true "name"
// @Success     200 {object} []interface{}
// @Failure     404 {object} string
// @Failure     500 {object} string
// @Router      /azure/subscription/{name} [get]
func (obj *AzureHandlerHttp) GetSubscription(c *gin.Context) {
	subs := c.Param("name")

	if len(subs) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "resource group name not setted"})
		return
	}

	result, err := obj.Service.GetSubscription(context.Background(), subs, subs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, "not found")
		return
	}

	c.JSON(http.StatusAccepted, result)
}

package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/synera-br/golang-cloud-collector/internal/core/service"
	"github.com/synera-br/golang-cloud-collector/pkg/otelpkg"
)

type AzureHandlerHttpInterface interface {
	ListResources(c *gin.Context)
	FindByResourceGroup(c *gin.Context)
	FindByTag(c *gin.Context)
	GetSubscription(c *gin.Context)
}

type AzureHandlerHttp struct {
	Service service.AzureServiceInterface
	Tracer  *otelpkg.OtelPkgInstrument
}

func NewAzureHandlerHttp(svc service.AzureServiceInterface, otl *otelpkg.OtelPkgInstrument, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) AzureHandlerHttpInterface {

	azure := &AzureHandlerHttp{
		Service: svc,
		Tracer:  otl,
	}

	azure.handlers(routerGroup, middleware...)

	return azure
}

func (c *AzureHandlerHttp) handlers(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	routerGroup.GET("/azure", append(middlewareList, c.ListResources)...)
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
	ctx, span := obj.Tracer.Tracer.Start(c.Request.Context(), "AzureHandlerHttp.ListResources")
	defer span.End()

	result, err := obj.Service.ListResources(ctx)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.New("not found")})
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
	ctx, span := obj.Tracer.Tracer.Start(c.Request.Context(), "AzureHandlerHttp.ListResources")
	defer span.End()

	rsg := c.Param("name")

	if len(rsg) == 0 {
		span.RecordError(errors.New("resource group name not setted"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "resource group name not setted"})
		return
	}

	result, err := obj.Service.ListResourcesByResourceGroup(ctx, rsg)
	if err != nil {
		span.RecordError(err)
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
	ctx, span := obj.Tracer.Tracer.Start(c.Request.Context(), "AzureHandlerHttp.ListResources")
	defer span.End()

	if len(c.Request.URL.Query()) != 2 {
		span.RecordError(errors.New("key and value of tags not setted"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "key and value of tags not setted"})
		return
	}

	result, err := obj.Service.ListResourcesByTag(ctx, c.Request.URL.Query().Get("key"), c.Request.URL.Query().Get("value"))
	if err != nil {
		span.RecordError(err)
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
	ctx, span := obj.Tracer.Tracer.Start(c.Request.Context(), "AzureHandlerHttp.ListResources")
	defer span.End()

	subs := c.Param("name")

	if len(subs) == 0 {
		span.RecordError(errors.New("resource group name not setted"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "resource group name not setted"})
		return
	}

	result, err := obj.Service.GetSubscription(ctx, subs, subs)
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

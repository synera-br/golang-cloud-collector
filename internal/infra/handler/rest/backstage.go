package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/synera-br/golang-cloud-collector/internal/core/entity"
	"github.com/synera-br/golang-cloud-collector/internal/core/service"
	"github.com/synera-br/golang-cloud-collector/pkg/otelpkg"
)

type BackstageHandlerHttpInterface interface {
	TriggerSyncProvider(c *gin.Context)
}

type BackstageHandlerHttp struct {
	Service service.BackstageServiceInterface
	Tracer  *otelpkg.OtelPkgInstrument
}

func NewBackstageHandlerHttp(svc service.BackstageServiceInterface, otl *otelpkg.OtelPkgInstrument, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) BackstageHandlerHttpInterface {

	azure := &BackstageHandlerHttp{
		Service: svc,
		Tracer:  otl,
	}

	azure.handlers(routerGroup, middleware...)

	return azure
}

func (c *BackstageHandlerHttp) handlers(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	routerGroup.POST("/backstage", append(middlewareList, c.TriggerSyncProvider)...)
	routerGroup.GET("/backstage", append(middlewareList, c.GetAllKinds)...)
	routerGroup.GET("/backstage/:namespace/:kind/:name", append(middlewareList, c.GetKind)...)
}

// BackstageSyncProvider    godoc
// @Summary     sync providers
// @Tags        backstage
// @Accept       json
// @Produce     json
// @Description get all backstage register
// @Param provider        query string true "name of provider"
// @Param account        query string false "name of account to filter"
// @Param key        query string false "tag key to filter"
// @Param value        query string false "tag value to filter"
// @Success     200 {object} entity.Trigger
// @Failure     404 {object} string
// @Failure     500 {object} string
// @Router      /backstage [post]
func (obj *BackstageHandlerHttp) TriggerSyncProvider(c *gin.Context) {
	ctx, span := obj.Tracer.Tracer.Start(c.Request.Context(), "BackstageHandlerHttp.TriggerSyncProvider")
	span.IsRecording()

	defer span.End()

	var trigger *entity.Trigger
	if err := c.ShouldBindJSON(&trigger); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := obj.Service.TriggerSyncProvider(ctx, trigger)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, response)
}

// BackstageGetAllKinds    godoc
// @Summary     kind all kinds
// @Tags        backstage
// @Accept       json
// @Produce     json
// @Param name        query string false "filter resource by name"
// @Param kind        query string false "filter resource by kind"
// @Param namespace        query string false "filter resource by namespace"
// @Description get all backstage register
// @Success     200 {object} []entity.KindReource
// @Failure     404 {object} string
// @Failure     500 {object} string
// @Router      /backstage [get]
func (obj *BackstageHandlerHttp) GetAllKinds(c *gin.Context) {
	ctx, span := obj.Tracer.Tracer.Start(c.Request.Context(), "BackstageHandlerHttp.GetAllKinds")
	defer span.End()

	result, err := obj.Service.GetAllKinds(ctx, entity.FilterKind{
		Name:      c.Request.URL.Query().Get("name"),
		Kind:      c.Request.URL.Query().Get("kind"),
		Namespace: c.Request.URL.Query().Get("namespace"),
	})

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

// BackstageGetKind    godoc
// @Summary     get specific kind
// @Tags        backstage
// @Accept       json
// @Produce     json
// @Param       namespace path string true "namespace of the resource"
// @Param       kind path string true "kind of the resource"
// @Param       name path string true "name of the resource"
// @Description get all backstage register
// @Success     200 {object} entity.KindReource
// @Failure     404 {object} string
// @Failure     500 {object} string
// @Router      /backstage/{namespace}/{kind}/{name} [get]
func (obj *BackstageHandlerHttp) GetKind(c *gin.Context) {
	ctx, span := obj.Tracer.Tracer.Start(c.Request.Context(), "BackstageHandlerHttp.GetKind")
	defer span.End()

	if c.Param("name") == "" || c.Param("kind") == "" || c.Param("namespace") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fields are empty"})
		return
	}

	filter := entity.FilterKind{
		Name:      c.Param("name"),
		Kind:      c.Param("kind"),
		Namespace: c.Param("namespace"),
	}

	result, err := obj.Service.GetAllKinds(ctx, filter)

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

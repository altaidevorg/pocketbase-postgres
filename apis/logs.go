package apis

import (
	"net/http"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/pocketbase/pocketbase/tools/search"
)

// bindLogsApi registers the request logs api endpoints.
func bindLogsApi(app core.App, rg *router.RouterGroup[*core.RequestEvent]) {
	sub := rg.Group("/logs").Bind(RequireSuperuserAuth(), SkipSuccessActivityLog())
	sub.GET("", logsList)
	sub.GET("/stats", logsStats)
	sub.GET("/{id}", logsView)
}

var logFilterFields = []string{
	"id", "created", "level", "message", "data",
	`^data\.[\w\.\:]*\w+$`,
}

func logsList(e *core.RequestEvent) error {
	fieldResolver := search.NewSimpleFieldResolver(logFilterFields...).SetIsPostgres(e.App.IsPostgres())

	// fallback to "created" sort if "@rowid" is used (usually by the Admin UI)
	// because "@rowid" is a special SQLite column and not available in other drivers
	params := e.Request.URL.Query()
	params.Set("sort", strings.ReplaceAll(params.Get("sort"), "@rowid", "created"))

	result, err := search.NewProvider(fieldResolver).
		Query(e.App.AuxModelQuery(&core.Log{})).
		ParseAndExec(params.Encode(), &[]*core.Log{})

	if err != nil {
		return e.BadRequestError("", err)
	}

	return e.JSON(http.StatusOK, result)
}

func logsStats(e *core.RequestEvent) error {
	fieldResolver := search.NewSimpleFieldResolver(logFilterFields...).SetIsPostgres(e.App.IsPostgres())

	filter := e.Request.URL.Query().Get(search.FilterQueryParam)

	var expr dbx.Expression
	if filter != "" {
		var err error
		expr, err = search.FilterData(filter).BuildExpr(fieldResolver)
		if err != nil {
			return e.BadRequestError("Invalid filter format.", err)
		}
	}

	stats, err := e.App.LogsStats(expr)
	if err != nil {
		return e.BadRequestError("Failed to generate logs stats.", err)
	}

	return e.JSON(http.StatusOK, stats)
}

func logsView(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if id == "" {
		return e.NotFoundError("", nil)
	}

	log, err := e.App.FindLogById(id)
	if err != nil || log == nil {
		return e.NotFoundError("", err)
	}

	return e.JSON(http.StatusOK, log)
}

package toolhandler

// BuiltinConfig 是内置处理器的可选配置。
type BuiltinConfig struct {
	WebSearchEndpoint string
	WebSearchAPIKey   string
	SQLDSN            string
}

// RegisterBuiltins 注册全部内置处理器，名称分别为：
// calculator、http_request、web_search、sql_query。
func RegisterBuiltins(r *Registry, cfg BuiltinConfig) {
	if r == nil {
		return
	}
	r.Register(NewCalculatorHandler())
	r.Register(NewHTTPHandler())
	r.Register(NewWebSearchHandler(cfg.WebSearchEndpoint, cfg.WebSearchAPIKey))
	r.Register(NewSQLQueryHandler(cfg.SQLDSN))
}

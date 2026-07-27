package pluginhost

var globalCatalog = NewCatalog()
var globalReloader *Reloader
var globalMarketService *MarketService

func SetGlobalCatalog(c *Catalog) {
	if c != nil {
		globalCatalog = c
	}
}

func GlobalCatalog() *Catalog {
	return globalCatalog
}

func SetGlobalReloader(r *Reloader) {
	globalReloader = r
}

func GlobalReloader() *Reloader {
	return globalReloader
}

func SetGlobalMarketService(s *MarketService) {
	globalMarketService = s
}

func GlobalMarketService() *MarketService {
	return globalMarketService
}

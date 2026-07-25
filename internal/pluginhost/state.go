package pluginhost

var globalCatalog = NewCatalog()
var globalReloader *Reloader

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

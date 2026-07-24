package pluginhost

var globalCatalog = NewCatalog()

func SetGlobalCatalog(c *Catalog) {
	if c != nil {
		globalCatalog = c
	}
}

func GlobalCatalog() *Catalog {
	return globalCatalog
}

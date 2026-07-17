package browser

func ListProxiesWithFallback(proxyDAO ProxyDAO, fallback []Proxy) []Proxy {
	if proxyDAO != nil {
		if list, err := proxyDAO.List(); err == nil {
			return list
		}
	}
	return append([]Proxy{}, fallback...)
}

func ListProxyGroups(proxyDAO ProxyDAO) []string {
	if proxyDAO != nil {
		if groups, err := proxyDAO.ListGroups(); err == nil {
			return groups
		}
	}
	return nil
}

func LatestProxiesWithFallback(proxyDAO ProxyDAO, fallback []Proxy) []Proxy {
	if proxyDAO != nil {
		if list, err := proxyDAO.List(); err == nil && len(list) > 0 {
			return list
		}
	}
	return fallback
}

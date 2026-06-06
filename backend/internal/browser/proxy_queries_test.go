package browser

import (
	"errors"
	"testing"
)

type proxyQueryTestDAO struct {
	list         []Proxy
	groups       []string
	groupList    []Proxy
	listErr      error
	groupsErr    error
	groupListErr error
}

func (d proxyQueryTestDAO) List() ([]Proxy, error)                              { return d.list, d.listErr }
func (d proxyQueryTestDAO) ListByGroup(string) ([]Proxy, error)                 { return d.groupList, d.groupListErr }
func (d proxyQueryTestDAO) ListGroups() ([]string, error)                       { return d.groups, d.groupsErr }
func (d proxyQueryTestDAO) Upsert(Proxy) error                                  { return nil }
func (d proxyQueryTestDAO) Delete(string) error                                 { return nil }
func (d proxyQueryTestDAO) DeleteAll() error                                    { return nil }
func (d proxyQueryTestDAO) UpdateSpeedResult(string, bool, int64, string) error { return nil }
func (d proxyQueryTestDAO) UpdateIPHealthResult(string, string) error           { return nil }

func TestListProxiesWithFallbackUsesDAO(t *testing.T) {
	fallback := []Proxy{{ProxyId: "fallback"}}
	list := ListProxiesWithFallback(proxyQueryTestDAO{list: []Proxy{{ProxyId: "dao"}}}, fallback)
	if len(list) != 1 || list[0].ProxyId != "dao" {
		t.Fatalf("list = %#v", list)
	}
}

func TestListProxiesWithFallbackCopiesFallback(t *testing.T) {
	fallback := []Proxy{{ProxyId: "fallback"}}
	list := ListProxiesWithFallback(proxyQueryTestDAO{listErr: errors.New("failed")}, fallback)
	list[0].ProxyId = "changed"
	if fallback[0].ProxyId != "fallback" {
		t.Fatalf("fallback was mutated")
	}
}

func TestListProxiesByGroupWithFallbackFiltersFallback(t *testing.T) {
	list := ListProxiesByGroupWithFallback(nil, "group-a", []Proxy{
		{ProxyId: "a", GroupName: "group-a"},
		{ProxyId: "b", GroupName: "group-b"},
	})
	if len(list) != 1 || list[0].ProxyId != "a" {
		t.Fatalf("list = %#v", list)
	}
}

func TestLatestProxiesWithFallbackKeepsFallbackForEmptyDAOList(t *testing.T) {
	fallback := []Proxy{{ProxyId: "fallback"}}
	list := LatestProxiesWithFallback(proxyQueryTestDAO{}, fallback)
	if len(list) != 1 || list[0].ProxyId != "fallback" {
		t.Fatalf("list = %#v", list)
	}
}

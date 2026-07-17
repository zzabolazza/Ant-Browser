package browser

import (
	"database/sql"
	"fmt"
	"time"
)

// ProxyDAO 代理列表持久化接口
type ProxyDAO interface {
	List() ([]Proxy, error)
	ListGroups() ([]string, error)
	Upsert(proxy Proxy) error
	Delete(proxyId string) error
	DeleteAll() error
	UpdateSpeedResult(proxyId string, ok bool, latencyMs int64, testedAt string) error
	UpdateIPHealthResult(proxyId string, healthJSON string) error
}

// SQLiteProxyDAO 基于 SQLite 的 ProxyDAO 实现
type SQLiteProxyDAO struct {
	db *sql.DB
}

// NewSQLiteProxyDAO 创建 SQLiteProxyDAO
func NewSQLiteProxyDAO(db *sql.DB) *SQLiteProxyDAO {
	return &SQLiteProxyDAO{db: db}
}

func (d *SQLiteProxyDAO) List() ([]Proxy, error) {
	rows, err := d.db.Query(`
		SELECT proxy_id, proxy_name, proxy_config, COALESCE(group_name, ''),
		       COALESCE(last_latency_ms, -1), COALESCE(last_test_ok, 0), COALESCE(last_tested_at, ''),
		       COALESCE(last_ip_health_json, ''),
		       sort_order
		FROM browser_proxies ORDER BY sort_order ASC, created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("查询代理列表失败: %w", err)
	}
	defer rows.Close()
	return scanProxies(rows)
}

func (d *SQLiteProxyDAO) ListGroups() ([]string, error) {
	rows, err := d.db.Query(`
		SELECT DISTINCT group_name FROM browser_proxies
		WHERE group_name != '' ORDER BY group_name ASC`)
	if err != nil {
		return nil, fmt.Errorf("查询代理分组失败: %w", err)
	}
	defer rows.Close()

	var groups []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (d *SQLiteProxyDAO) Upsert(proxy Proxy) error {
	now := time.Now().Format(time.RFC3339)
	_, err := d.db.Exec(`
		INSERT INTO browser_proxies (
		  proxy_id, proxy_name, proxy_config, group_name,
		  sort_order, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(proxy_id) DO UPDATE SET
		  proxy_name   = excluded.proxy_name,
		  proxy_config = excluded.proxy_config,
		  group_name   = excluded.group_name,
		  sort_order   = excluded.sort_order`,
		proxy.ProxyId, proxy.ProxyName, proxy.ProxyConfig, proxy.GroupName,
		proxy.SortOrder, now,
	)
	if err != nil {
		return fmt.Errorf("保存代理失败: %w", err)
	}
	return nil
}

func (d *SQLiteProxyDAO) Delete(proxyId string) error {
	_, err := d.db.Exec(`DELETE FROM browser_proxies WHERE proxy_id = ?`, proxyId)
	if err != nil {
		return fmt.Errorf("删除代理失败: %w", err)
	}
	return nil
}

func (d *SQLiteProxyDAO) DeleteAll() error {
	_, err := d.db.Exec(`DELETE FROM browser_proxies`)
	if err != nil {
		return fmt.Errorf("清空代理表失败: %w", err)
	}
	return nil
}

func (d *SQLiteProxyDAO) UpdateSpeedResult(proxyId string, ok bool, latencyMs int64, testedAt string) error {
	okInt := 0
	if ok {
		okInt = 1
	}
	_, err := d.db.Exec(`
		UPDATE browser_proxies SET last_latency_ms=?, last_test_ok=?, last_tested_at=?
		WHERE proxy_id=?`, latencyMs, okInt, testedAt, proxyId)
	if err != nil {
		return fmt.Errorf("更新测速结果失败: %w", err)
	}
	return nil
}

func (d *SQLiteProxyDAO) UpdateIPHealthResult(proxyId string, healthJSON string) error {
	_, err := d.db.Exec(`
		UPDATE browser_proxies SET last_ip_health_json=?
		WHERE proxy_id=?`, healthJSON, proxyId)
	if err != nil {
		return fmt.Errorf("更新 IP 健康结果失败: %w", err)
	}
	return nil
}

func scanProxies(rows *sql.Rows) ([]Proxy, error) {
	var list []Proxy
	for rows.Next() {
		var p Proxy
		var okInt int
		if err := rows.Scan(
			&p.ProxyId, &p.ProxyName, &p.ProxyConfig, &p.GroupName,
			&p.LastLatencyMs, &okInt, &p.LastTestedAt, &p.LastIPHealthJSON, &p.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("读取代理行失败: %w", err)
		}
		p.LastTestOk = okInt == 1
		list = append(list, p)
	}
	return list, rows.Err()
}

package backend

import (
	"database/sql"
	"fmt"
	"strings"
)

type backupMergeColumn struct {
	name     string
	fallback string
	expr     string
}

// backupBuildCompatibleMergeSQL keeps merge imports compatible with snapshots
// created before optional columns were introduced by later migrations.
func backupBuildCompatibleMergeSQL(tx *sql.Tx, table string, resetFirst bool) (string, bool, error) {
	columns, err := backupSrcColumns(tx, table)
	if err != nil {
		return "", false, err
	}
	where := ""
	var targetColumns []backupMergeColumn

	switch table {
	case "browser_groups":
		parentExpr := backupMergedGroupParentExpr("s")
		targetColumns = []backupMergeColumn{
			{name: "group_id", fallback: "''"}, {name: "group_name", fallback: "''"},
			{name: "parent_id", fallback: "''", expr: parentExpr}, {name: "sort_order", fallback: "0"},
			{name: "created_at", fallback: "CURRENT_TIMESTAMP"}, {name: "updated_at", fallback: "CURRENT_TIMESTAMP"},
		}
		where = fmt.Sprintf(`NOT EXISTS (
  SELECT 1 FROM browser_groups t
  WHERE t.group_id = s.group_id
     OR (t.parent_id = %s AND lower(t.group_name) = lower(s.group_name))
)
AND NOT EXISTS (
  SELECT 1 FROM src.browser_groups earlier
  WHERE earlier.rowid < s.rowid
    AND (earlier.group_id = s.group_id
      OR (earlier.parent_id = s.parent_id AND lower(earlier.group_name) = lower(s.group_name)))
)`, parentExpr)
	case "browser_proxies":
		targetColumns = []backupMergeColumn{
			{name: "proxy_id", fallback: "''"}, {name: "proxy_name", fallback: "''"},
			{name: "proxy_config", fallback: "''"}, {name: "group_name", fallback: "''"},
			{name: "last_latency_ms", fallback: "-1"},
			{name: "last_test_ok", fallback: "0"}, {name: "last_tested_at", fallback: "''"},
			{name: "last_ip_health_json", fallback: "''"},
			{name: "sort_order", fallback: "0"}, {name: "created_at", fallback: "CURRENT_TIMESTAMP"},
		}
		where = `NOT EXISTS (
  SELECT 1 FROM browser_proxies t
  WHERE (trim(s.proxy_id) <> '' AND t.proxy_id = s.proxy_id)
     OR (trim(s.proxy_config) <> '' AND lower(t.proxy_config) = lower(s.proxy_config))
)
AND NOT EXISTS (
  SELECT 1 FROM src.browser_proxies earlier
  WHERE earlier.rowid < s.rowid
    AND ((trim(s.proxy_id) <> '' AND earlier.proxy_id = s.proxy_id)
      OR (trim(s.proxy_config) <> '' AND lower(earlier.proxy_config) = lower(s.proxy_config)))
)`
	case "browser_profiles":
		targetColumns = []backupMergeColumn{
			{name: "profile_id", fallback: "''"}, {name: "profile_name", fallback: "''"},
			{name: "user_data_dir", fallback: "''"},
			{name: "core_id", fallback: "''", expr: backupMergedProfileCoreExpr(columns)},
			{name: "fingerprint_args", fallback: "'[]'"},
			{name: "proxy_id", fallback: "''", expr: backupMergedProfileProxyExpr(columns)},
			{name: "proxy_config", fallback: "''"}, {name: "proxy_bind_name", fallback: "''"},
			{name: "proxy_bind_updated_at", fallback: "''"}, {name: "launch_args", fallback: "'[]'"},
			{name: "tags", fallback: "'[]'"}, {name: "keywords", fallback: "'[]'"},
			{name: "group_id", fallback: "''", expr: backupMergedProfileGroupExpr(columns)},
			{name: "created_at", fallback: "CURRENT_TIMESTAMP"}, {name: "updated_at", fallback: "CURRENT_TIMESTAMP"},
		}
		where = `NOT EXISTS (
  SELECT 1 FROM browser_profiles t
  WHERE t.profile_id = s.profile_id
     OR (trim(s.user_data_dir) <> '' AND lower(t.user_data_dir) = lower(s.user_data_dir))
)
AND NOT EXISTS (
  SELECT 1 FROM src.browser_profiles earlier
  WHERE earlier.rowid < s.rowid
    AND (earlier.profile_id = s.profile_id
      OR (trim(s.user_data_dir) <> '' AND lower(earlier.user_data_dir) = lower(s.user_data_dir)))
)`
		if _, hasDeletedAt := columns["deleted_at"]; hasDeletedAt {
			where += "\nAND trim(COALESCE(s.deleted_at, '')) = ''"
		}
	case "browser_extensions":
		targetColumns = []backupMergeColumn{
			{name: "extension_id", fallback: "''"}, {name: "name", fallback: "''"},
			{name: "version", fallback: "''"}, {name: "description", fallback: "''"},
			{name: "icon_data_url", fallback: "''"}, {name: "manifest_json", fallback: "'{}'"},
			{name: "source_url", fallback: "''"}, {name: "install_dir", fallback: "''"},
			{name: "enabled", fallback: "1"}, {name: "installed_at", fallback: "''"},
			{name: "updated_at", fallback: "''"},
		}
		where = `NOT EXISTS (SELECT 1 FROM browser_extensions t WHERE t.extension_id = s.extension_id)`
	default:
		return "", false, nil
	}

	insertColumns := make([]string, 0, len(targetColumns))
	selectColumns := make([]string, 0, len(targetColumns))
	for _, column := range targetColumns {
		insertColumns = append(insertColumns, quoteBackupIdentifier(column.name))
		expr := column.expr
		if expr == "" {
			expr = backupSrcValue(columns, column.name, column.fallback)
		}
		selectColumns = append(selectColumns, expr)
	}
	query := fmt.Sprintf("INSERT INTO %s (%s)\nSELECT %s\nFROM src.%s s",
		quoteBackupIdentifier(table), strings.Join(insertColumns, ", "), strings.Join(selectColumns, ", "), quoteBackupIdentifier(table))
	if !resetFirst && where != "" {
		query += "\nWHERE " + where
	}
	return query, true, nil
}

func backupMergedGroupParentExpr(sourceAlias string) string {
	parentID := sourceAlias + ".parent_id"
	return fmt.Sprintf(`CASE WHEN trim(COALESCE(%[1]s, '')) = '' THEN ''
  WHEN EXISTS (SELECT 1 FROM browser_groups direct_parent WHERE direct_parent.group_id = %[1]s) THEN %[1]s
  ELSE COALESCE((
    SELECT target_parent.group_id
    FROM src.browser_groups source_parent
    JOIN browser_groups target_parent ON lower(target_parent.group_name) = lower(source_parent.group_name)
    WHERE source_parent.group_id = %[1]s
    ORDER BY CASE WHEN target_parent.parent_id = source_parent.parent_id THEN 0 ELSE 1 END, target_parent.rowid
    LIMIT 1
  ), %[1]s) END`, parentID)
}

func backupSrcColumns(tx *sql.Tx, table string) (map[string]struct{}, error) {
	rows, err := tx.Query("PRAGMA src.table_info(" + quoteBackupIdentifier(table) + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := map[string]struct{}{}
	for rows.Next() {
		var cid, notNull, pk int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns[strings.ToLower(name)] = struct{}{}
	}
	return columns, rows.Err()
}

func backupSrcValue(columns map[string]struct{}, name, fallback string) string {
	if _, ok := columns[strings.ToLower(name)]; !ok {
		return fallback
	}
	return "COALESCE(s." + quoteBackupIdentifier(name) + ", " + fallback + ")"
}

func quoteBackupIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func backupMergedProfileCoreExpr(columns map[string]struct{}) string {
	if _, ok := columns["core_id"]; !ok {
		return "''"
	}
	return `CASE WHEN trim(COALESCE(s.core_id, '')) = '' THEN ''
  WHEN EXISTS (SELECT 1 FROM browser_cores t WHERE t.core_id = s.core_id) THEN s.core_id
  ELSE COALESCE((SELECT t.core_id FROM browser_cores t JOIN src.browser_cores sc ON lower(t.core_path) = lower(sc.core_path) WHERE sc.core_id = s.core_id LIMIT 1), '') END`
}

func backupMergedProfileProxyExpr(columns map[string]struct{}) string {
	if _, ok := columns["proxy_id"]; !ok {
		return "''"
	}
	return `CASE WHEN trim(COALESCE(s.proxy_id, '')) = '' THEN ''
  WHEN EXISTS (SELECT 1 FROM browser_proxies t WHERE t.proxy_id = s.proxy_id) THEN s.proxy_id
  ELSE COALESCE((SELECT t.proxy_id FROM browser_proxies t JOIN src.browser_proxies sp ON lower(t.proxy_config) = lower(sp.proxy_config) WHERE sp.proxy_id = s.proxy_id LIMIT 1), '') END`
}

func backupMergedProfileGroupExpr(columns map[string]struct{}) string {
	if _, ok := columns["group_id"]; !ok {
		return "''"
	}
	return `CASE WHEN trim(COALESCE(s.group_id, '')) = '' THEN ''
  WHEN EXISTS (SELECT 1 FROM browser_groups t WHERE t.group_id = s.group_id) THEN s.group_id
  ELSE COALESCE((SELECT t.group_id FROM browser_groups t JOIN src.browser_groups sg ON lower(t.group_name) = lower(sg.group_name) WHERE sg.group_id = s.group_id LIMIT 1), '') END`
}

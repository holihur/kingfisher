package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RunMigrations 执行 dir 目录下所有 *.up.sql 迁移文件（按文件名排序，幂等）。
//
// 原理：
//  1. 创建追踪表 schema_migrations(version PK, applied_at)
//  2. 读取 dir 下 *.up.sql，按字典序排序
//  3. 已在追踪表的 version 跳过，否则读取文件内容并在事务中执行，成功后写入追踪表
//
// dir 不存在或为空目录返回 error，便于调用方发现配置错误。
// 文件内的多语句按 `;` 切分逐条执行，兼容 MySQL multiStatements 未开启的情况。
func RunMigrations(db *gorm.DB, dir string) error {
	// 1. 检查目录
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("migrations dir %q: %w", dir, err)
	}

	// 2. 确保追踪表存在
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	// 3. 已应用的版本
	applied, err := loadApplied(db)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	// 4. 收集待执行的 *.up.sql
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, name)
		}
	}
	if len(files) == 0 {
		return fmt.Errorf("no *.up.sql found in %q", dir)
	}
	sort.Strings(files)

	// 5. 逐文件执行
	for _, name := range files {
		version := strings.TrimSuffix(name, ".up.sql") // 如 000001_create_users
		if applied[version] {
			continue
		}
		fullPath := filepath.Join(dir, name)
		content, err := os.ReadFile(fullPath) // #nosec G304 -- dir 与 name 均来自受控的 migrations 目录
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		sqlText := strings.TrimSpace(string(content))
		if sqlText == "" {
			// 空文件也标记为已应用，避免反复尝试
			if err := recordApplied(db, version); err != nil {
				return err
			}
			continue
		}
		if err := execSQLFile(db, sqlText); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if err := recordApplied(db, version); err != nil {
			return fmt.Errorf("record %s: %w", name, err)
		}
	}
	return nil
}

func ensureMigrationsTable(db *gorm.DB) error {
	// 兼容 MySQL / Postgres / SQLite 的建表语句（用 TEXT + DATETIME 通用类型）
	ddl := `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at DATETIME NOT NULL
)`
	return db.Exec(ddl).Error
}

func loadApplied(db *gorm.DB) (map[string]bool, error) {
	rows, err := db.Raw("SELECT version FROM schema_migrations").Rows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	m := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		m[v] = true
	}
	return m, rows.Err()
}

func recordApplied(db *gorm.DB, version string) error {
	// 使用 Exec 而非 Create，避免依赖模型定义
	return db.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", version, time.Now().UTC()).Error
}

func execSQLFile(db *gorm.DB, content string) error {
	// 按 ; 切分，多语句逐条执行；单事务保证原子性
	return db.Transaction(func(tx *gorm.DB) error {
		for _, stmt := range splitSQL(content) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			// 去掉单行注释前缀的空行已由 Trim 处理
			if strings.HasPrefix(stmt, "--") {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// splitSQL 按分号切分 SQL，保留引号内的分号不切（简化版：按 ; 分割后过滤空语句即可满足本项目 migrations）。
func splitSQL(content string) []string {
	// 本项目 migrations 均为简单 DDL/DML，无存储过程，不含引号内分号，按 ; 切分足够
	return strings.Split(content, ";")
}

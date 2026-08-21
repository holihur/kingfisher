#!/bin/bash
# Start E2E backend — 支持多后端（Go / Java），通过 E2E_BACKEND 环境变量切换
# 用法: E2E_BACKEND=go bash scripts/e2e-server.sh   # 默认 Go
#       E2E_BACKEND=java bash scripts/e2e-server.sh # Java Spring Boot
# 脚本由 playwright 以 repo root 为 cwd 执行（webServer.cwd: ../../），用相对路径定位，兼容 CI
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

# 方案选择：默认 go，支持 E2E_BACKEND / BACKEND 两种环境变量名
BACKEND="${E2E_BACKEND:-${BACKEND:-go}}"
BACKEND=$(echo "$BACKEND" | tr '[:upper:]' '[:lower:]')

# 统一清理：DB + Redis 状态（Go/Java 共用 e2e 库与 Redis db=1）
rm -f /tmp/kingfisher-e2e.db
redis-cli -n 1 keys "login_fail:*" 2>/dev/null | xargs -r redis-cli -n 1 del 2>/dev/null
redis-cli -n 1 keys "user:perms:*" 2>/dev/null | xargs -r redis-cli -n 1 del 2>/dev/null
redis-cli -n 1 keys "menu:*" 2>/dev/null | xargs -r redis-cli -n 1 del 2>/dev/null
# 清理 asynq 任务队列（e2e.yaml 启用了 taskqueue，避免残留任务串扰用例）
redis-cli -n 1 --scan --pattern 'asynq:*' 2>/dev/null | xargs -r redis-cli -n 1 del 2>/dev/null
# Java 额外清理：内存回退的黑名单/限流（已随 FLUSHDB 清理，保留扩展）
redis-cli -n 1 keys "blacklist:*" 2>/dev/null | xargs -r redis-cli -n 1 del 2>/dev/null

if [ "$BACKEND" = "java" ]; then
  echo "[e2e-server] Backend: Java (Spring Boot) on :18080, DB=/tmp/kingfisher-e2e.db, Redis db=1"
  # 加载 .env（与 KingfisherApplication.loadDotEnv 对齐，支持 export 前缀）
  set -a; [ -f .env ] && . ./.env; [ -f java/.env ] && . ./java/.env; set +a
  # e2e 专用 JWT 与端口（与 config/e2e.yaml 对齐）
  export KINGFISHER_DB="/tmp/kingfisher-e2e.db"
  export JWT_SECRET="${JWT_SECRET:-e2e-test-secret-jwt-key-2024}"
  # 若无 jar 则先编译（CI 首次）
  if [ ! -f "java/target/kingfisher-0.0.1-SNAPSHOT.jar" ]; then
    echo "[e2e-server] Java jar 不存在，先编译..."
    (cd java && mvn package -DskipTests -q)
  fi
  # 用 Go 先播种一次（创建表 + 种子用户），再切 Java（Java 无 AutoMigrate）
  echo "[e2e-server] 播种数据库..."
  if [ -f "kingfisher.db" ]; then
    echo "[e2e-server] 复制现有 kingfisher.db -> /tmp/kingfisher-e2e.db"
    cp kingfisher.db /tmp/kingfisher-e2e.db
  else
    echo "[e2e-server] kingfisher.db 不存在，用 Go 播种..."
    export CONFIG_PATH=config/e2e.yaml
    timeout 20 bash -c 'go run ./cmd/server/ 2>&1 | head -n 20' || true
    # 等待 DB 文件生成且包含 users 表
    for i in 1 2 3 4 5; do
      if sqlite3 /tmp/kingfisher-e2e.db "SELECT count(*) FROM users" 2>/dev/null | grep -qE "[0-9]+"; then
        echo "[e2e-server] 播种完成"
        break
      fi
      sleep 1
    done
  fi
  # 兜底：若仍为空则强制用 Go 同步播种一次（较慢但可靠）
  if ! sqlite3 /tmp/kingfisher-e2e.db "SELECT count(*) FROM users" >/dev/null 2>&1; then
    echo "[e2e-server] 兜底播种..."
    export CONFIG_PATH=config/e2e.yaml
    timeout 15 go run ./cmd/server/ &
    GO_PID=$!
    for i in 1 2 3 4 5 6 7 8 9 10; do
      sleep 1
      if sqlite3 /tmp/kingfisher-e2e.db "SELECT count(*) FROM users" 2>/dev/null | grep -qE "[0-9]+"; then break; fi
    done
    kill $GO_PID 2>/dev/null || true
    wait $GO_PID 2>/dev/null || true
  fi
  # 启动 Java（与 config/e2e.yaml 关键配置对齐：port 18080, issuer, ttl, redis db 1, log level error）
  exec java -jar java/target/kingfisher-0.0.1-SNAPSHOT.jar \
    --server.port=18080 \
    --spring.datasource.url=jdbc:sqlite:/tmp/kingfisher-e2e.db \
    --spring.data.redis.database=1 \
    --spring.data.redis.host=127.0.0.1 \
    --spring.data.redis.port=6379 \
    --kingfisher.jwt.secret="${JWT_SECRET}" \
    --kingfisher.jwt.issuer=kingfisher-e2e \
    --kingfisher.jwt.access-ttl=3600h \
    --kingfisher.jwt.refresh-ttl=7200h \
    --logging.level.com.kingfisher=ERROR \
    --logging.level.org.springframework.data.redis=WARN
else
  echo "[e2e-server] Backend: Go on :18080, DB=/tmp/kingfisher-e2e.db, Redis db=1"
  export CONFIG_PATH=config/e2e.yaml
  exec go run ./cmd/server/
fi

#!/usr/bin/env bash
# Kingfisher 二进制部署脚本（不使用 Docker）
# 流程：本机已 push → SSH 到目标机 → git pull/clone → 构建前后端 → systemd 重启 → 健康检查
#
# 用法：bash scripts/deploy.sh   （或 task deploy）
# 环境变量可覆盖：
#   DEPLOY_HOST=47.95.200.101  DEPLOY_PORT=10022  DEPLOY_USER=root
#   DEPLOY_DIR=/root/kingfisher  DEPLOY_SERVICE=kingfisher
#   DEPLOY_SSH_KEY=~/.ssh/kingfisher_deploy（登录用 key，默认 ~/.ssh/kingfisher_deploy）
set -euo pipefail

HOST="${DEPLOY_HOST:-47.95.200.101}"
SSHPORT="${DEPLOY_PORT:-10022}"
USER="${DEPLOY_USER:-root}"
DIR="${DEPLOY_DIR:-/root/kingfisher}"
SERVICE="${DEPLOY_SERVICE:-kingfisher}"
KEY="${DEPLOY_SSH_KEY:-$HOME/.ssh/kingfisher_deploy}"
REPO="git@github.com:kingfisher-vvv/kingfisher.git"

SSH="ssh -p $SSHPORT -o ConnectTimeout=10 -o BatchMode=yes -o StrictHostKeyChecking=accept-new -i $KEY"

echo "==> [1/6] 本机构建前端 dist（随后端静态服务提供）"
(cd "$(dirname "$0")/../kingfisher-web" && corepack enable && pnpm install --frozen-lockfile >/dev/null && pnpm build >/dev/null)
echo "    前端 dist 构建完成"

echo "==> [2/6] 上传前端产物到 ${USER}@${HOST}:${DIR}"
# 首次需要 git clone；之后 git pull。前端 dist 通过 git 仓库带上（含在仓库内）
$SSH "$USER@$HOST" "mkdir -p $DIR"
$SSH "$USER@$HOST" "if [ ! -d $DIR/.git ]; then git clone $REPO $DIR; else cd $DIR && git pull --ff-only; fi"
# 前端 dist 已 gitignore，需本机上传（避免构建环境依赖目标机 node）
scp -P "$SSHPORT" -i "$KEY" -r "$(dirname "$0")/../kingfisher-web/dist" "$USER@$HOST:$DIR/kingfisher-web/"

echo "==> [3/6] 生成生产配置 + systemd unit"
$SSH "$USER@$HOST" "mkdir -p $DIR/config $DIR/logs $DIR/bin"
# 生产配置：端口 8090、release 模式；JWT_SECRET 通过 systemd 注入
scp -P "$SSHPORT" -i "$KEY" "$(dirname "$0")/../deploy/config.prod.yaml" "$USER@$HOST:$DIR/config/config.yaml"
scp -P "$SSHPORT" -i "$KEY" "$(dirname "$0")/../deploy/kingfisher.service" "$USER@$HOST:/etc/systemd/system/$SERVICE.service"

echo "==> [4/6] 目标机构建后端二进制"
$SSH "$USER@$HOST" "cd $DIR && CGO_ENABLED=0 go build -ldflags='-s -w -X main.version=\$(git describe --tags --always 2>/dev/null || echo dev) -X main.commit=\$(git rev-parse --short HEAD) -X main.buildTime=\$(date -u +%Y-%m-%dT%H:%M:%SZ)' -o bin/server ./cmd/server"

echo "==> [5/6] 重启 systemd 服务"
$SSH "$USER@$HOST" "systemctl daemon-reload && systemctl enable $SERVICE >/dev/null 2>&1; systemctl restart $SERVICE"

echo "==> [6/6] 健康检查 http://${HOST}:8090"
for i in $(seq 1 15); do
  if curl -sf "http://$HOST:8090/health" >/dev/null 2>&1; then
    echo "    部署成功 ✓  http://$HOST:8090  （health OK，$i 秒）"
    exit 0
  fi
  sleep 2
done
echo "    部署失败：服务未在 15 次探测内就绪" >&2
$SSH "$USER@$HOST" "systemctl status $SERVICE --no-pager -l | tail -20" || true
exit 1

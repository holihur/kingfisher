#!/usr/bin/env bash
# Kingfisher 部署脚本：从 GitHub Release 拉取 CI 构建产物到 47.95.200.101，systemd 运行。
# 不在目标机编译，直接拉取二进制 + 前端 dist（仓库公开，免 token）。
#
# 用法：bash scripts/deploy.sh   （或 task deploy）
# 环境变量可覆盖：
#   DEPLOY_HOST=47.95.200.101  DEPLOY_PORT=10022  DEPLOY_USER=root
#   DEPLOY_DIR=/root/kingfisher  DEPLOY_SERVICE=kingfisher  DEPLOY_VERSION=<tag 或 latest>
#   DEPLOY_SSH_KEY=~/.ssh/kingfisher_deploy
set -euo pipefail

HOST="${DEPLOY_HOST:-47.95.200.101}"
SSHPORT="${DEPLOY_PORT:-10022}"
USER="${DEPLOY_USER:-root}"
DIR="${DEPLOY_DIR:-/root/kingfisher}"
SERVICE="${DEPLOY_SERVICE:-kingfisher}"
KEY="${DEPLOY_SSH_KEY:-$HOME/.ssh/kingfisher}"
VERSION="${DEPLOY_VERSION:-latest}"   # 默认拉最新 release
REPO="kingfisher-vvv/kingfisher"
ASSET="kingfisher-deploy.tar.gz"

SSH="ssh -p $SSHPORT -o ConnectTimeout=10 -o BatchMode=yes -o StrictHostKeyChecking=accept-new -i $KEY"

echo "==> [1/5] 获取 GitHub Release 资产下载地址（${REPO} @ ${VERSION}）"
if [ "$VERSION" = "latest" ]; then
  ASSET_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
else
  ASSET_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
fi

echo "==> [2/5] 在目标机拉取并解压产物到 ${DIR}"
$SSH "$USER@$HOST" "set -e; mkdir -p $DIR/release && cd $DIR/release && \
  curl -sfL '$ASSET_URL' -o ${ASSET} && \
  tar -xzf ${ASSET} && \
  echo '  解压完成:'; ls -la"

echo "==> [3/5] 应用产物（备份旧版本 + 放置新版本）"
$SSH "$USER@$HOST" "set -e; cd $DIR && \
  mkdir -p logs && \
  # 备份当前运行版本
  if [ -f kingfisher-server ]; then cp kingfisher-server kingfisher-server.bak; fi
  # 新产物进主目录
  cp -r release/kingfisher-server . 2>/dev/null || cp release/kingfisher-server .
  rm -rf kingfisher-web && cp -r release/kingfisher-web .
  cp release/config.yaml config.yaml
  chmod +x kingfisher-server
  # 确保 JWT secret 未用占位值（生产安全）
  grep -q 'please-change-me-in-production' /etc/systemd/system/${SERVICE}.service && echo '  ⚠ 请先修改 JWT_SECRET！' || true"

echo "==> [4/5] 安装 systemd unit 并重启"
# systemd unit 放在仓库 deploy/ 下，首次由脚本拷贝；之后如无变更可跳过
scp -P "$SSHPORT" -i "$KEY" "$(dirname "$0")/../deploy/kingfisher.service" "$USER@$HOST:/etc/systemd/system/${SERVICE}.service"
$SSH "$USER@$HOST" "systemctl daemon-reload && systemctl enable ${SERVICE} >/dev/null 2>&1; systemctl restart ${SERVICE}"

echo "==> [5/5] 健康检查 http://${HOST}:8090"
for i in $(seq 1 15); do
  if curl -sf "http://$HOST:8090/health" >/dev/null 2>&1; then
    echo "    部署成功 ✓  http://$HOST:8090  （health OK，${i} 秒）"
    # 清理旧的备份（保留 1 份）
    $SSH "$USER@$HOST" "cd $DIR && rm -f kingfisher-server.bak.bak" || true
    exit 0
  fi
  sleep 2
done
echo "    部署失败：服务未在 15 次探测内就绪" >&2
$SSH "$USER@$HOST" "systemctl status ${SERVICE} --no-pager -l | tail -25" || true
exit 1

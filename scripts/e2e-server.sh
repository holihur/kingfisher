#!/bin/bash
# Start E2E backend — cleanup DB and Redis state first
rm -f /tmp/kingfisher-e2e.db
redis-cli -n 1 keys "login_fail:*" 2>/dev/null | xargs -r redis-cli -n 1 del 2>/dev/null
redis-cli -n 1 keys "user:perms:*" 2>/dev/null | xargs -r redis-cli -n 1 del 2>/dev/null
redis-cli -n 1 keys "menu:*" 2>/dev/null | xargs -r redis-cli -n 1 del 2>/dev/null
export CONFIG_PATH=config/e2e.yaml
exec go run /Users/jzx/Desktop/kingfisher/cmd/server/

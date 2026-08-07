import { execSync } from 'child_process';

async function globalSetup() {
  // Clear login fail counters from previous test runs
  try {
    execSync('redis-cli keys "login_fail:*" | xargs -r redis-cli del', { stdio: 'ignore' });
  } catch {
    // Redis might not be available — ignore
  }
}

export default globalSetup;

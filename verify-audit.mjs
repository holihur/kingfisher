import { chromium } from '@playwright/test';
const b = await chromium.launch({ channel: 'chrome', headless: true });
const p = await b.newPage();
const r = await p.request.post('http://localhost:8080/api/v1/auth/login', { data: { username: 'admin', password: 'Abcd1234' } });
const body = await r.json();
await p.goto('http://localhost:5173/login');
await p.evaluate(({t,rf}) => { localStorage.setItem('kingfisher_token', t); localStorage.setItem('kingfisher_refresh', rf); }, { t: body.data.access_token, rf: body.data.refresh_token });
await p.goto('http://localhost:5173/system/audit', { waitUntil: 'networkidle' });
await p.waitForTimeout(2000);
// 表格有内容
const rowCount = await p.locator('.ant-table-row').count();
console.log('审计行数:', rowCount);
// 中文操作 Tag
console.log('页面含"创建"Tag:', await p.locator('.ant-table').innerText().then(t => t.includes('创建')));
console.log('页面含"更新"Tag:', await p.locator('.ant-table').innerText().then(t => t.includes('更新')));
console.log('页面含中文资源"系统配置":', await p.locator('.ant-table').innerText().then(t => t.includes('系统配置')));
// 点击查看详情
const viewBtn = p.locator('a', { hasText: '查看' }).first();
if (await viewBtn.count()) {
  await viewBtn.click();
  await p.waitForTimeout(600);
  console.log('详情 Modal 打开:', await p.locator('.ant-modal').isVisible());
  console.log('Modal 含详情 JSON:', await p.locator('.ant-modal').innerText().then(t => t.includes('"value"')));
}
await b.close();

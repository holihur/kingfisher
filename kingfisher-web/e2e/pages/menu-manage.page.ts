import type { Page, Locator } from '@playwright/test';
import { URLS } from '../utils/constants';

/**
 * Page Object Model for the Menu Manage page.
 */
export class MenuManagePage {
  constructor(readonly page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto(URLS.menus);
    await this.page.waitForLoadState('networkidle');
  }

  addRootButton(): Locator {
    return this.page.getByRole('button', { name: '新增根菜单' });
  }

  addChildLink(rowText: string): Locator {
    return this.page.locator(`tr`, { hasText: rowText }).getByText('添加子项');
  }

  editLink(rowText: string): Locator {
    return this.page.locator(`tr`, { hasText: rowText }).getByText('编辑');
  }

  deleteLink(rowText: string): Locator {
    return this.page.locator(`tr`, { hasText: rowText }).getByText('删除');
  }

  table(): Locator {
    return this.page.locator('.ant-table');
  }

  tableRows(): Locator {
    return this.table().locator('.ant-table-row');
  }

  // Modal
  modal(): Locator {
    return this.page.locator('.ant-modal');
  }

  modalTitle(): Locator {
    return this.modal().locator('.ant-modal-title');
  }

  // Type tags
  typeTag(text: string): Locator {
    return this.page.locator('.ant-tag').filter({ hasText: text });
  }

  popconfirm(): Locator {
    return this.page.locator('.ant-popconfirm');
  }

  popconfirmOk(): Locator {
    return this.popconfirm().getByRole('button', { name: '确 定' });
  }
}

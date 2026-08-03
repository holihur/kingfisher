import type { Page, Locator } from '@playwright/test';
import { URLS } from '../utils/constants';

/**
 * Page Object Model for the User List page.
 */
export class UserListPage {
  constructor(readonly page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto(URLS.users);
    await this.page.waitForLoadState('networkidle');
  }

  addButton(): Locator {
    return this.page.getByRole('button', { name: '新增用户' });
  }

  searchInput(): Locator {
    return this.page.getByPlaceholder('搜索');
  }

  searchButton(): Locator {
    return this.page.getByRole('button', { name: '搜索' });
  }

  resetButton(): Locator {
    return this.page.getByRole('button', { name: '重置' });
  }

  table(): Locator {
    return this.page.locator('.ant-table');
  }

  tableRows(): Locator {
    return this.table().locator('.ant-table-row');
  }

  editLink(username: string): Locator {
    return this.page.locator(`tr`, { hasText: username }).getByText('编辑');
  }

  deleteLink(username: string): Locator {
    return this.page.locator(`tr`, { hasText: username }).getByText('删除');
  }

  // Modal
  modal(): Locator {
    return this.page.locator('.ant-modal');
  }

  modalTitle(): Locator {
    return this.modal().locator('.ant-modal-title');
  }

  modalSubmitButton(): Locator {
    return this.modal().getByRole('button', { name: /确定|保存/ });
  }

  modalCancelButton(): Locator {
    return this.modal().getByRole('button', { name: '取消' });
  }

  // Form fields inside modal
  usernameField(): Locator {
    return this.modal().locator('#username');
  }

  passwordField(): Locator {
    return this.modal().locator('#password');
  }

  emailField(): Locator {
    return this.modal().locator('#email');
  }

  statusSelect(): Locator {
    return this.modal().locator('#status');
  }

  roleSelect(): Locator {
    return this.modal().locator('#role_id');
  }

  // Popconfirm
  popconfirm(): Locator {
    return this.page.locator('.ant-popconfirm');
  }

  popconfirmOk(): Locator {
    return this.popconfirm().getByRole('button', { name: '确 定' });
  }

  pagination(): Locator {
    return this.page.locator('.ant-pagination');
  }

  async search(keyword: string): Promise<void> {
    await this.searchInput().fill(keyword);
    await this.searchButton().click();
    await this.page.waitForLoadState('networkidle');
  }

  async resetSearch(): Promise<void> {
    await this.resetButton().click();
    await this.page.waitForLoadState('networkidle');
  }

  async gotoPage(pageNum: number): Promise<void> {
    await this.pagination().getByText(String(pageNum), { exact: true }).click();
    await this.page.waitForLoadState('networkidle');
  }
}

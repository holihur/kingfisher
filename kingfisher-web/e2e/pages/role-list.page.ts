import type { Page, Locator } from '@playwright/test';
import { URLS } from '../utils/constants';

/**
 * Page Object Model for the Role List page.
 */
export class RoleListPage {
  constructor(readonly page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto(URLS.roles);
    await this.page.waitForLoadState('networkidle');
  }

  addButton(): Locator {
    return this.page.getByRole('button', { name: '新增角色' });
  }

  permissionsLink(roleName: string): Locator {
    return this.page.locator(`tr`, { hasText: roleName }).getByText('权限');
  }

  menusLink(roleName: string): Locator {
    return this.page.locator(`tr`, { hasText: roleName }).getByText('菜单');
  }

  editLink(roleName: string): Locator {
    return this.page.locator(`tr`, { hasText: roleName }).getByText('编辑');
  }

  deleteLink(roleName: string): Locator {
    return this.page.locator(`tr`, { hasText: roleName }).getByText('删除');
  }

  table(): Locator {
    return this.page.locator('.ant-table');
  }

  tableRows(): Locator {
    return this.table().locator('.ant-table-row');
  }

  modal(): Locator {
    return this.page.locator('.ant-modal');
  }

  modalTitle(): Locator {
    return this.modal().locator('.ant-modal-title');
  }

  popconfirm(): Locator {
    return this.page.locator('.ant-popconfirm');
  }

  popconfirmOk(): Locator {
    return this.popconfirm().getByRole('button', { name: '确 定' });
  }
}

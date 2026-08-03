import type { Page, Locator } from '@playwright/test';
import { URLS } from '../utils/constants';

/**
 * Page Object Model for the Config Manage page.
 */
export class ConfigManagePage {
  constructor(readonly page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto(URLS.configs);
    await this.page.waitForLoadState('networkidle');
  }

  editLink(configKey: string): Locator {
    return this.page.locator(`tr`, { hasText: configKey }).getByText('编辑');
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

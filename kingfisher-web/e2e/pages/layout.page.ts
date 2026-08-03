import type { Page, Locator } from '@playwright/test';

/**
 * Page Object Model for the Admin Layout (sidebar, header, breadcrumb).
 */
export class LayoutPage {
  constructor(readonly page: Page) {}

  logo(): Locator {
    return this.page.locator('.ant-layout-sider').getByText('Kingfisher');
  }

  collapsedLogo(): Locator {
    return this.page.locator('.ant-layout-sider').getByText('K', { exact: true });
  }

  collapseButton(): Locator {
    return this.page.locator('.ant-layout-sider .anticon').first();
  }

  sidebarMenu(): Locator {
    return this.page.locator('.ant-menu-root');
  }

  menuItem(label: string): Locator {
    return this.sidebarMenu().getByText(label, { exact: true });
  }

  submenuItem(group: string, item: string): Locator {
    return this.sidebarMenu().getByText(item);
  }

  headerUser(): Locator {
    return this.page.locator('.ant-layout-header').locator('span').filter({ hasText: /.+/ });
  }

  avatar(): Locator {
    return this.page.locator('.ant-layout-header .ant-avatar');
  }

  breadcrumb(): Locator {
    return this.page.locator('.ant-breadcrumb');
  }

  logoutMenuItem(): Locator {
    return this.page.getByText('退出登录');
  }

  contentArea(): Locator {
    return this.page.locator('.ant-layout-content');
  }

  async collapse(): Promise<void> {
    await this.collapseButton().click();
    await this.page.waitForTimeout(300); // animation
  }

  async expand(): Promise<void> {
    await this.collapseButton().click();
    await this.page.waitForTimeout(300);
  }

  async logout(): Promise<void> {
    // Click the header user area to open dropdown
    const headerSpan = this.page.locator('.ant-layout-header').getByText(/.+/).first();
    await headerSpan.click();
    await this.logoutMenuItem().click();
    await this.page.waitForURL('**/login');
  }

  async navigateTo(label: string): Promise<void> {
    await this.sidebarMenu().getByText(label).click();
  }
}

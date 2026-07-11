import { expect, type Locator, type Page } from "@host-tests/support/playwright";

import { LoginPage } from "@host-tests/pages/LoginPage";
import { PluginPage } from "@host-tests/pages/PluginPage";

export class DemoSourcePage extends PluginPage {
  constructor(page: Page) {
    super(page);
  }

  get pluginLoginSlot(): Locator {
    return this.page.getByText(
      "linapro-demo-source 已向登录页公开区注册扩展内容，用于验证 `auth.login.after` 插槽。",
    );
  }

  async goto() {
    await new LoginPage(this.page).goto();
  }

  async loginAndWaitForRedirect(username: string, password: string) {
    await new LoginPage(this.page).loginAndWaitForRedirect(username, password);
  }

  headerActionBeforeSlot(): Locator {
    return this.page.getByText("linapro-demo-source 头部前置扩展").first();
  }

  headerActionAfterSlot(): Locator {
    return this.page.getByText("linapro-demo-source 头部后置扩展").first();
  }

  pluginSidebarIntroTitle(): Locator {
    return this.page
      .getByRole("heading", { name: "源码插件示例已生效" })
      .first();
  }

  pluginSidebarIntroSummary(): Locator {
    return this.page.getByText(
      "这是一条来自 linapro-demo-source 接口的简要介绍，用于验证源码插件菜单页可读取插件后端数据。",
    );
  }

  pluginSourceRecordGridTitle(): Locator {
    return this.page.getByText("示例记录").first();
  }

  pluginSourceRecordAddButton(): Locator {
    return this.page.getByTestId("linapro-demo-source-record-add").first();
  }

  pluginSourceRecordModal(): Locator {
    return this.page
      .getByRole("dialog", { name: /新增示例记录|编辑示例记录/ })
      .last();
  }

  pluginSourceRecordAttachmentAlert(): Locator {
    return this.page
      .getByTestId("linapro-demo-source-record-attachment-alert")
      .last();
  }

  pluginSourceRecordUploadSection(): Locator {
    return this.page
      .getByTestId("linapro-demo-source-record-upload-section")
      .last();
  }

  pluginSourceRecordExistingAttachment(): Locator {
    return this.page
      .getByTestId("linapro-demo-source-record-existing-attachment")
      .last();
  }

  pluginSourceRecordRemoveAttachmentOption(): Locator {
    return this.page
      .getByTestId("linapro-demo-source-record-remove-attachment-option")
      .last();
  }

  pluginSourceRecordDragger(): Locator {
    return this.page.getByTestId("linapro-demo-source-record-dragger").last();
  }

  pluginSourceRecordTitleInput(): Locator {
    return this.page
      .getByTestId("linapro-demo-source-record-title-input")
      .last();
  }

  pluginSourceRecordContentInput(): Locator {
    return this.page
      .getByTestId("linapro-demo-source-record-content-input")
      .last();
  }

  pluginSourceRecordRow(title: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: title }).first();
  }

  workspaceBeforeSlot(): Locator {
    return this.page.getByText(
      "linapro-demo-source 正在通过 `dashboard.workspace.before` 在工作台顶部插入横幅内容。",
    );
  }

  workspaceAfterSlot(): Locator {
    return this.page.getByText("源码插件示例工作台卡片").first();
  }

  crudToolbarSlot(): Locator {
    return this.page.getByText("linapro-demo-source CRUD 扩展").first();
  }

  async expectWorkspaceSlotHidden() {
    await expect(this.workspaceBeforeSlot()).toHaveCount(0);
    await expect(this.workspaceAfterSlot()).toHaveCount(0);
  }

  async expectHeaderSlotsHidden() {
    await expect(this.headerActionBeforeSlot()).toHaveCount(0);
    await expect(this.headerActionAfterSlot()).toHaveCount(0);
  }

  async expectCrudSlotsHidden() {
    await expect(this.crudToolbarSlot()).toHaveCount(0);
  }

  async openSidebarExampleFromMenu() {
    await this.clickSidebarMenuItem("示例插件-源码插件");
    await expect(this.pluginSidebarIntroTitle()).toHaveCount(0);
    await expect(this.pluginSidebarIntroSummary()).toHaveCount(0);
    await expect(this.pluginSourceRecordGridTitle()).toBeVisible();
  }

  async createSourceDemoRecord(
    title: string,
    content: string,
    filePath?: string,
  ) {
    await expect(this.pluginSourceRecordAddButton()).toBeVisible();
    await this.pluginSourceRecordAddButton().click();
    await expect(this.pluginSourceRecordModal()).toBeVisible();
    await this.expectSourceRecordModalCompactLayout();
    await this.pluginSourceRecordTitleInput().fill(title);
    await this.pluginSourceRecordContentInput().fill(content);
    if (filePath) {
      const [fileChooser] = await Promise.all([
        this.page.waitForEvent("filechooser"),
        this.pluginSourceRecordDragger().click(),
      ]);
      await fileChooser.setFiles(filePath);
      await expect(
        this.pluginSourceRecordModal().locator(".ant-upload-list-item"),
      ).toBeVisible();
    }
    await this.pluginSourceRecordModal()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(this.pluginSourceRecordModal()).toHaveCount(0);
    await expect(this.pluginSourceRecordRow(title)).toBeVisible();
  }

  async editSourceDemoRecord(
    currentTitle: string,
    nextTitle: string,
    nextContent: string,
  ) {
    const editButton = await this.pluginSourceRecordActionButton(
      currentTitle,
      /编\s*辑/,
    );
    await expect(editButton).toBeVisible();
    await editButton.click();
    await expect(this.pluginSourceRecordModal()).toBeVisible();
    await expect(this.pluginSourceRecordTitleInput()).toHaveValue(currentTitle);
    await this.expectSourceRecordModalCompactLayout();
    await this.pluginSourceRecordTitleInput().fill(nextTitle);
    await this.pluginSourceRecordContentInput().fill(nextContent);
    await this.pluginSourceRecordModal()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(this.pluginSourceRecordModal()).toHaveCount(0);
    await expect(this.pluginSourceRecordRow(nextTitle)).toBeVisible();
  }

  async deleteSourceDemoRecord(title: string) {
    const deleteButton = await this.pluginSourceRecordActionButton(
      title,
      /删\s*除/,
    );
    await expect(deleteButton).toBeVisible();
    await deleteButton.click();
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
    await expect(this.pluginSourceRecordRow(title)).toHaveCount(0);
  }

  async downloadSourceDemoAttachment(fileName: string) {
    const downloadPromise = this.page.waitForEvent("download");
    await this.page.getByRole("button", { name: fileName }).first().click();
    return await downloadPromise;
  }

  private async pluginSourceRecordActionButton(title: string, name: RegExp) {
    const row = this.pluginSourceRecordRow(title);
    await expect(row, `未找到示例记录行: ${title}`).toBeVisible();
    return row.getByRole("button", { name }).first();
  }

  private async expectSourceRecordModalCompactLayout() {
    const modal = this.pluginSourceRecordModal();
    const alert = this.pluginSourceRecordAttachmentAlert();
    const uploadSection = this.pluginSourceRecordUploadSection();

    await expect(alert).toBeVisible();
    await expect(uploadSection).toBeVisible();

    const modalWidth = await modal.evaluate((element) => {
      return Math.round(element.getBoundingClientRect().width);
    });
    expect(
      modalWidth,
      "源码插件记录弹窗宽度应收敛，避免继续维持过宽布局",
    ).toBeLessThanOrEqual(620);

    const alertBox = await alert.boundingBox();
    const uploadSectionBox = await uploadSection.boundingBox();
    expect(alertBox, "附件提示块应可见").toBeTruthy();
    expect(uploadSectionBox, "上传区域应可见").toBeTruthy();

    const verticalGap = uploadSectionBox!.y - (alertBox!.y + alertBox!.height);
    expect(
      verticalGap,
      "附件提示块与上传区域之间应保留至少 16px 的垂直间距",
    ).toBeGreaterThanOrEqual(16);

    const existingAttachment = this.pluginSourceRecordExistingAttachment();
    const removeAttachmentOption =
      this.pluginSourceRecordRemoveAttachmentOption();
    const editSpacingVisible = await existingAttachment
      .isVisible({ timeout: 1000 })
      .catch(() => false);
    if (!editSpacingVisible) {
      return;
    }

    await expect(removeAttachmentOption).toBeVisible();
    const existingAttachmentBox = await existingAttachment.boundingBox();
    const removeAttachmentBox = await removeAttachmentOption.boundingBox();
    const draggerBox = await this.pluginSourceRecordDragger().boundingBox();
    expect(existingAttachmentBox, "当前附件信息块应可见").toBeTruthy();
    expect(removeAttachmentBox, "移除附件选项块应可见").toBeTruthy();
    expect(draggerBox, "附件上传区应可见").toBeTruthy();

    const removeOptionGapAbove =
      removeAttachmentBox!.y -
      (existingAttachmentBox!.y + existingAttachmentBox!.height);
    const removeOptionGapBelow =
      draggerBox!.y - (removeAttachmentBox!.y + removeAttachmentBox!.height);
    expect(
      removeOptionGapAbove,
      "“提交时移除当前附件”选项与当前附件信息块之间应保留足够间距",
    ).toBeGreaterThanOrEqual(12);
    expect(
      removeOptionGapBelow,
      "“提交时移除当前附件”选项与上传区之间应保留足够间距",
    ).toBeGreaterThanOrEqual(12);
  }
}

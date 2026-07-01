import type { Page } from '@playwright/test';

import { mkdir } from 'node:fs/promises';
import path from 'node:path';

import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { PostPage } from '../../pages/PostPage';

async function captureEvidence(page: Page, name: string) {
  const now = new Date();
  const day = new Intl.DateTimeFormat('en-CA', {
    day: '2-digit',
    month: '2-digit',
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
  })
    .format(now)
    .replaceAll('-', '');
  const time = new Intl.DateTimeFormat('en-GB', {
    hour: '2-digit',
    hour12: false,
    minute: '2-digit',
    second: '2-digit',
    timeZone: 'Asia/Shanghai',
  })
    .format(now)
    .replaceAll(':', '');
  const dir = path.resolve(process.cwd(), '..', '..', 'temp', day);
  await mkdir(dir, { recursive: true });
  await page.screenshot({
    fullPage: true,
    path: path.join(dir, `${time}-${name}.png`),
  });
}

test.describe('TC004 岗位新增后部门树刷新展开', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'linapro-org-core');
  });

  test('TC004a: 连续新增岗位后左侧部门树仍保持展开', async ({
    adminPage,
  }) => {
    const postPage = new PostPage(adminPage);
    const parentDept = 'LinaPro.AI';
    const childDept = '研发部门';
    const suffix = Date.now();
    const posts = [
      {
        code: `TREE_REFRESH_${suffix}_A`,
        name: `树刷新岗位A_${suffix}`,
      },
      {
        code: `TREE_REFRESH_${suffix}_B`,
        name: `树刷新岗位B_${suffix}`,
      },
    ];
    const createdCodes: string[] = [];

    await postPage.goto();
    expect(await postPage.hasVisibleDeptNode(childDept)).toBeTruthy();
    await captureEvidence(adminPage, 'post-dept-tree-initial');

    try {
      for (const [index, post] of posts.entries()) {
        await postPage.createPost(parentDept, post.code, post.name);
        createdCodes.push(post.code);

        expect(await postPage.hasVisibleDeptNode(childDept)).toBeTruthy();
        await captureEvidence(
          adminPage,
          `post-dept-tree-after-create-${index + 1}`,
        );
      }
    } finally {
      for (const code of createdCodes.reverse()) {
        await postPage.deletePost(code).catch(() => {});
      }
    }
  });
});

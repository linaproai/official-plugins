<script setup lang="ts">
import type { Dept, DeptTree } from './dept-client';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { getPopupContainer } from '@vben/utils';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { $t } from '#/locales';
import {
  deptAdd,
  deptExclude,
  deptInfo,
  deptTree,
  deptUpdate,
  deptUsers,
} from './dept-client';
import { useDictStore } from '#/store/dict';

import { drawerSchema } from './dept-data';

const emit = defineEmits<{ reload: [] }>();

const dictStore = useDictStore();

interface DrawerProps {
  id?: number;
  update: boolean;
}

type DeptSelectSource = Partial<Omit<DeptTree, 'children'>> &
  Pick<Dept, 'id'> & {
  children?: DeptSelectSource[];
  name?: string;
  parentId?: number;
};

type DeptSelectNode = Omit<DeptTree, 'children'> & {
  children?: DeptSelectNode[];
  fullName?: string;
  parentId?: number;
};

const TOP_LEVEL_DEPT_ID = 0;
const TOP_LEVEL_DEPT_LABEL_KEY = 'plugin.linapro-org-core.dept.tree.topLevelDept';
const isUpdate = ref(false);
const deptId = ref<number>(0);
const title = computed(() =>
  isUpdate.value
    ? $t('plugin.linapro-org-core.dept.drawer.editTitle')
    : $t('plugin.linapro-org-core.dept.drawer.createTitle'),
);

const [BasicForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-2',
    labelWidth: 80,
  },
  schema: drawerSchema(),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

/** 为树节点添加 fullName（显示完整路径） */
function addFullName(
  tree: DeptSelectNode[],
  parentPath = '',
  separator = ' / ',
) {
  for (const node of tree) {
    const fullName = parentPath
      ? `${parentPath}${separator}${node.label}`
      : node.label;
    (node as any).fullName = fullName;
    if (node.children && node.children.length > 0) {
      addFullName(node.children, fullName, separator);
    }
  }
}

function normalizeDeptSelectTree(items: DeptSelectSource[]): DeptSelectNode[] {
  if (items.some((item) => typeof item.parentId === 'number')) {
    return buildDeptSelectTree(items);
  }
  return items.map((item) => ({
    id: item.id,
    label: item.label ?? item.name ?? '',
    children: normalizeDeptSelectTree(item.children ?? []),
  }));
}

function buildDeptSelectTree(items: DeptSelectSource[]): DeptSelectNode[] {
  const nodeMap = new Map<number, DeptSelectNode>();
  for (const item of items) {
    nodeMap.set(item.id, {
      id: item.id,
      label: item.label ?? item.name ?? '',
      parentId: item.parentId,
      children: [],
    });
  }

  const roots: DeptSelectNode[] = [];
  for (const item of items) {
    const node = nodeMap.get(item.id);
    if (!node) {
      continue;
    }
    const parentId = item.parentId ?? TOP_LEVEL_DEPT_ID;
    const parent = parentId === TOP_LEVEL_DEPT_ID
      ? undefined
      : nodeMap.get(parentId);
    if (parent) {
      parent.children.push(node);
    } else {
      roots.push(node);
    }
  }
  return roots;
}

function topLevelDeptLabel() {
  const translated = $t(TOP_LEVEL_DEPT_LABEL_KEY);
  if (translated !== TOP_LEVEL_DEPT_LABEL_KEY) {
    return translated;
  }
  return preferences.app.locale === 'en-US' ? 'Top-level Department' : '顶级部门';
}

/** 初始化部门树选择 */
async function initDeptSelect(id?: number) {
  let rawTreeData: DeptSelectSource[];
  if (isUpdate.value && id) {
    // 编辑时排除自身及子节点
    rawTreeData = ((await deptExclude(id)) || []) as DeptSelectSource[];
  } else {
    rawTreeData = ((await deptTree()) || []) as DeptSelectSource[];
  }
  const selectableTree: DeptSelectNode[] = [
    {
      id: TOP_LEVEL_DEPT_ID,
      label: topLevelDeptLabel(),
      children: normalizeDeptSelectTree(rawTreeData),
    },
  ];
  // 添加完整路径名
  addFullName(selectableTree);
  formApi.updateSchema([
    {
      componentProps: {
        fieldNames: { label: 'label', value: 'id' },
        getPopupContainer,
        showSearch: true,
        treeData: selectableTree,
        treeDefaultExpandAll: true,
        treeDefaultExpandedKeys: [TOP_LEVEL_DEPT_ID],
        treeLine: { showLeafIcon: false },
        treeNodeFilterProp: 'label',
        treeNodeLabelProp: 'fullName',
      },
      fieldName: 'parentId',
    },
  ]);
}

/** 加载负责人用户列表 */
async function loadLeaderUsers(targetDeptId: number, keyword?: string) {
  const ret = await deptUsers(targetDeptId, { keyword, limit: 10 });
  const options = ret.map((user) => ({
    label: `${user.username} | ${user.nickname}`,
    value: Number(user.id),
  }));
  formApi.updateSchema([
    {
      componentProps: {
        filterOption: false,
        onSearch: (val: string) => loadLeaderUsers(targetDeptId, val),
        options,
        placeholder: $t('plugin.linapro-org-core.dept.placeholders.selectLeader'),
        showSearch: true,
      },
      fieldName: 'leader',
    },
  ]);
}

/** 初始化部门负责人下拉 */
async function initDeptUsers(targetDeptId: number) {
  await loadLeaderUsers(targetDeptId);
}

const [BasicDrawer, drawerApi] = useVbenDrawer({
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  async onOpenChange(isOpen) {
    if (!isOpen) {
      return;
    }
    drawerApi.setState({ loading: true });

    const { id, update } = drawerApi.getData() as DrawerProps;
    isUpdate.value = update;

    if (id) {
      await formApi.setFieldValue('parentId', id);
      if (update) {
        deptId.value = id;
        const record = await deptInfo(id);
        // Convert leader=0 to undefined so the select shows blank
        if (record.leader === 0) {
          record.leader = undefined as any;
        }
        // Load options before setValues so the current leader can match a label.
        await initDeptUsers(id);
        await formApi.setValues(record);
      } else {
        await initDeptUsers(0);
      }
    } else {
      await formApi.setFieldValue('parentId', TOP_LEVEL_DEPT_ID);
      await initDeptUsers(0);
    }

    await initDeptSelect(id);

    // 加载字典：状态选项
    const statusOptions = await dictStore.getDictOptionsAsync('sys_normal_disable');
    formApi.updateSchema([
      {
        fieldName: 'status',
        componentProps: {
          options: statusOptions.map((d) => ({
            label: d.label,
            value: Number(d.value),
          })),
        },
      },
    ]);

    drawerApi.setState({ loading: false });
  },
});

async function handleConfirm() {
  try {
    drawerApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const data = await formApi.getValues();
    if (isUpdate.value) {
      await deptUpdate(deptId.value, data);
      message.success($t('pages.common.updateSuccess'));
    } else {
      await deptAdd(data);
      message.success($t('pages.common.createSuccess'));
    }
    emit('reload');
    drawerApi.close();
  } catch (error) {
    const messageText =
      error instanceof Error
        ? error.message
        : $t('plugin.linapro-org-core.dept.messages.saveFailed');
    message.error(messageText);
  } finally {
    drawerApi.lock(false);
  }
}

async function handleClosed() {
  await formApi.resetForm();
  deptId.value = 0;
}
</script>

<template>
  <BasicDrawer :title="title" class="w-[600px]">
    <BasicForm />
  </BasicDrawer>
</template>

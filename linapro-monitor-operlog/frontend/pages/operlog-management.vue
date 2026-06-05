<script lang="ts">
export const pluginPageMeta = {
  routePath: '/monitor/operlog',
  title: 'Audit Logs',
};
</script>

<script setup lang="ts">
import type { OperLog } from './operlog-client';

import { onMounted, ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';

import {
  Alert,
  Checkbox,
  DatePicker,
  message,
  Modal,
  Space,
} from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  operLogClean,
  operLogExport,
  operLogList,
} from './operlog-client';
import { $t } from '#/locales';
import { downloadBlob } from '#/utils/download';
import { useDictStore } from '#/store/dict';

import { buildColumns, buildQuerySchema } from './data';
import OperlogDetailDrawer from './operlog-detail-drawer.vue';

const dictStore = useDictStore();
const RangePicker = DatePicker.RangePicker;

onMounted(async () => {
  // Wait for dictionary requests to finish before wiring select options into the form.
  const [operTypeOptions, operStatusOptions] = await Promise.all([
    dictStore.getDictOptionsAsync('sys_oper_type'),
    dictStore.getDictOptionsAsync('sys_oper_status'),
  ]);
  gridApi.formApi.updateSchema([
    {
      fieldName: 'operType',
      componentProps: {
        options: operTypeOptions.map((d: any) => ({
          label: d.label,
          value: d.value,
        })),
      },
    },
    {
      fieldName: 'status',
      componentProps: {
        options: operStatusOptions.map((d: any) => ({
          label: d.label,
          value: d.value,
        })),
      },
    },
  ]);
});

const [DetailDrawerRef, detailDrawerApi] = useVbenDrawer({
  connectedComponent: OperlogDetailDrawer,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildQuerySchema(),
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  gridOptions: {
    columns: buildColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    sortConfig: {
      remote: true,
      trigger: 'cell',
    },
    proxyConfig: {
      sort: true,
      ajax: {
        query: async ({ page, sorts }: any, formValues: Record<string, any> = {}) => {
          const sortParams: Record<string, string> = {};
          if (sorts && sorts.length > 0) {
            const sort = sorts[0];
            if (sort && sort.order) {
              sortParams.orderBy = sort.field;
              sortParams.orderDirection = sort.order;
            }
          }

          const params: Record<string, any> = {
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
            ...sortParams,
          };

          // Handle operTime date range
          if (params.operTime && Array.isArray(params.operTime)) {
            params.beginTime = params.operTime[0];
            params.endTime = params.operTime[1];
            delete params.operTime;
          }

          return await operLogList(params);
        },
      },
    },
    headerCellConfig: {
      height: 44,
    },
    cellConfig: {
      height: 48,
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'linapro-monitor-operlog-index',
  },
});

const deleteRange = ref<string[]>([]);
const deleteAllLogs = ref(false);
const deleteRangeModalOpen = ref(false);
const deleteRangeSubmitting = ref(false);

function handlePreview(row: OperLog) {
  detailDrawerApi.setData({ record: row });
  detailDrawerApi.open();
}

function handleClean() {
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('plugin.linapro-monitor-operlog.messages.clearConfirm'),
    onOk: async () => {
      await operLogClean();
      message.success($t('plugin.linapro-monitor-operlog.messages.clearSuccess'));
      await gridApi.reload();
    },
  });
}

function handleDelete() {
  deleteRange.value = [];
  deleteAllLogs.value = false;
  deleteRangeModalOpen.value = true;
}

async function handleDeleteRangeConfirm() {
  const [beginTime, endTime] = deleteRange.value;
  if (!deleteAllLogs.value && (!beginTime || !endTime)) {
    message.warning(
      $t('plugin.linapro-monitor-operlog.messages.deleteRangeRequired'),
    );
    return;
  }

  deleteRangeSubmitting.value = true;
  try {
    const result = await operLogClean(
      deleteAllLogs.value ? undefined : { beginTime, endTime },
    );
    message.success(
      $t('plugin.linapro-monitor-operlog.messages.deleteRangeSuccess', {
        count: result?.deleted ?? 0,
      }),
    );
    deleteRangeModalOpen.value = false;
    await gridApi.query();
  } finally {
    deleteRangeSubmitting.value = false;
  }
}

async function handleExport() {
  const content = $t('pages.exportConfirm.all');

  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'primary',
    content,
    okText: $t('pages.common.confirm'),
    cancelText: $t('pages.common.cancel'),
    onOk: async () => {
      try {
        const formValues = gridApi.formApi.form.values;
        const params: Record<string, any> = { ...formValues };

        if (params.operTime && Array.isArray(params.operTime)) {
          params.beginTime = params.operTime[0];
          params.endTime = params.operTime[1];
          delete params.operTime;
        }

        const data = await operLogExport(params);
        downloadBlob(data, $t('plugin.linapro-monitor-operlog.exportFileName'));
        message.success($t('pages.common.exportSuccess'));
      } catch {
        message.error($t('pages.common.exportFailed'));
      }
    },
  });
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid :table-title="$t('plugin.linapro-monitor-operlog.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleClean">{{ $t('pages.common.clear') }}</a-button>
          <a-button @click="handleExport">{{ $t('pages.common.export') }}</a-button>
          <a-button
            data-testid="operlog-range-delete"
            danger
            type="primary"
            @click="handleDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <ghost-button @click.stop="handlePreview(row)">
          {{ $t('pages.common.detail') }}
        </ghost-button>
      </template>
    </Grid>

    <DetailDrawerRef />

    <Modal
      v-model:open="deleteRangeModalOpen"
      :destroy-on-close="true"
      :title="$t('plugin.linapro-monitor-operlog.messages.deleteRangeTitle')"
    >
      <div>
        <div data-testid="operlog-delete-alert">
          <Alert
            :message="$t('plugin.linapro-monitor-operlog.messages.deleteRangeDescription')"
            show-icon
            type="warning"
          />
        </div>
        <div data-testid="operlog-delete-all-option" style="margin-top: 16px">
          <Checkbox v-model:checked="deleteAllLogs">
            {{ $t('plugin.linapro-monitor-operlog.messages.deleteAllLabel') }}
          </Checkbox>
          <div class="text-xs text-gray-500" style="margin-top: 4px">
            {{ $t('plugin.linapro-monitor-operlog.messages.deleteAllHint') }}
          </div>
        </div>
        <div data-testid="operlog-delete-range-section" style="margin-top: 16px">
          <RangePicker
            v-model:value="deleteRange"
            :disabled="deleteAllLogs"
            class="w-full"
            value-format="YYYY-MM-DD"
          />
        </div>
      </div>
      <template #footer>
        <a-button @click="deleteRangeModalOpen = false">
          {{ $t('pages.common.cancel') }}
        </a-button>
        <a-button
          :loading="deleteRangeSubmitting"
          danger
          type="primary"
          @click="handleDeleteRangeConfirm"
        >
          {{ $t('pages.common.confirm') }}
        </a-button>
      </template>
    </Modal>
  </Page>
</template>

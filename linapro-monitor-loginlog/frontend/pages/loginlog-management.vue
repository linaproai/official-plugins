<script lang="ts">
export const pluginPageMeta = {
  routePath: '/monitor/loginlog',
  title: 'Login History',
};
</script>

<script setup lang="ts">
import type { LoginLog } from './loginlog-client';

import { onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

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
  loginLogClean,
  loginLogExport,
  loginLogList,
} from './loginlog-client';
import { $t } from '#/locales';
import { useDictStore } from '#/store/dict';
import { downloadBlob } from '#/utils/download';

import { buildColumns, buildQuerySchema } from './data';
import LoginlogDetailModal from './loginlog-detail-modal.vue';

const dictStore = useDictStore();
const RangePicker = DatePicker.RangePicker;

onMounted(async () => {
  const statusOptions = await dictStore.getDictOptionsAsync('sys_login_status');
  gridApi.formApi.updateSchema([
    {
      fieldName: 'status',
      componentProps: {
        options: statusOptions.map((d: any) => ({
          label: d.label,
          value: d.value,
        })),
      },
    },
  ]);
});

const [DetailModalRef, detailModalApi] = useVbenModal({
  connectedComponent: LoginlogDetailModal,
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
        query: async (
          { page, sorts }: any,
          formValues: Record<string, any> = {},
        ) => {
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

          if (params.loginTime && Array.isArray(params.loginTime)) {
            params.beginTime = params.loginTime[0];
            params.endTime = params.loginTime[1];
            delete params.loginTime;
          }

          return await loginLogList(params);
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
    id: 'linapro-monitor-loginlog-index',
  },
});

const deleteRange = ref<string[]>([]);
const deleteAllLogs = ref(false);
const deleteRangeModalOpen = ref(false);
const deleteRangeSubmitting = ref(false);

function handlePreview(row: LoginLog) {
  detailModalApi.setData(row);
  detailModalApi.open();
}

function handleClean() {
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('plugin.linapro-monitor-loginlog.messages.clearConfirm'),
    onOk: async () => {
      await loginLogClean();
      message.success($t('plugin.linapro-monitor-loginlog.messages.clearSuccess'));
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
      $t('plugin.linapro-monitor-loginlog.messages.deleteRangeRequired'),
    );
    return;
  }

  deleteRangeSubmitting.value = true;
  try {
    const result = await loginLogClean(
      deleteAllLogs.value ? undefined : { beginTime, endTime },
    );
    message.success(
      $t('plugin.linapro-monitor-loginlog.messages.deleteRangeSuccess', {
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

        if (params.loginTime && Array.isArray(params.loginTime)) {
          params.beginTime = params.loginTime[0];
          params.endTime = params.loginTime[1];
          delete params.loginTime;
        }

        const data = await loginLogExport(params);
        downloadBlob(data, $t('plugin.linapro-monitor-loginlog.exportFileName'));
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
    <Grid :table-title="$t('plugin.linapro-monitor-loginlog.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleClean">{{ $t('pages.common.clear') }}</a-button>
          <a-button @click="handleExport">{{ $t('pages.common.export') }}</a-button>
          <a-button
            data-testid="loginlog-range-delete"
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

    <DetailModalRef />

    <Modal
      v-model:open="deleteRangeModalOpen"
      :destroy-on-close="true"
      :title="$t('plugin.linapro-monitor-loginlog.messages.deleteRangeTitle')"
    >
      <div>
        <div data-testid="loginlog-delete-alert">
          <Alert
            :message="$t('plugin.linapro-monitor-loginlog.messages.deleteRangeDescription')"
            show-icon
            type="warning"
          />
        </div>
        <div data-testid="loginlog-delete-all-option" style="margin-top: 16px">
          <Checkbox v-model:checked="deleteAllLogs">
            {{ $t('plugin.linapro-monitor-loginlog.messages.deleteAllLabel') }}
          </Checkbox>
          <div class="text-xs text-gray-500" style="margin-top: 4px">
            {{ $t('plugin.linapro-monitor-loginlog.messages.deleteAllHint') }}
          </div>
        </div>
        <div data-testid="loginlog-delete-range-section" style="margin-top: 16px">
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

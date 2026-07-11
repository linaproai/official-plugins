<script lang="ts">
export const pluginPageMeta = {
  routePath: "/ai/invocations",
  title: "Request Logs",
};
</script>

<script setup lang="ts">
import type { Invocation } from "./ai-client";

import { ref } from "vue";
import dayjs from "dayjs";

import { Page, useVbenDrawer } from "@vben/common-ui";
import { useAccess } from "@vben/access";

import {
  Alert,
  Checkbox,
  DatePicker,
  message,
  Modal,
  Space,
} from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import { invocationClean, invocationList } from "./ai-client";
import {
  buildInvocationColumns,
  buildInvocationQuerySchema,
  splitCapabilityMethod,
} from "./ai-data";
import InvocationDetailDrawer from "./invocation-detail-drawer.vue";

const [DetailDrawerRef, detailDrawerApi] = useVbenDrawer({
  connectedComponent: InvocationDetailDrawer,
});

const accessCodes = {
  clear: "ai:invocation:clear",
};
const { hasAccessByCodes } = useAccess();
const RangePicker = DatePicker.RangePicker;

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildInvocationQuerySchema(),
    commonConfig: {
      labelWidth: 80,
      componentProps: { allowClear: true },
    },
    wrapperClass:
      "grid-cols-1 md:grid-cols-2 xl:grid-cols-[minmax(384px,1.45fr)_minmax(228px,1fr)_minmax(190px,0.95fr)_minmax(190px,0.95fr)] xl:gap-x-0",
  },
  gridOptions: {
    columns: buildInvocationColumns(),
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          const filters = normalizeInvocationFilters(formValues);
          return await invocationList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...filters,
          });
        },
      },
    },
    rowConfig: { keyField: "id" },
    id: "linapro-ai-core-invocation-index",
  },
});

const deleteRange = ref<string[]>([]);
const deleteAllLogs = ref(false);
const deleteRangeModalOpen = ref(false);
const deleteRangeSubmitting = ref(false);

function resolveRangeBoundary(value: unknown, boundary: "end" | "start") {
  if (value === undefined || value === null || value === "") {
    return undefined;
  }
  const rawValue = String(value).trim();
  if (!rawValue) {
    return undefined;
  }
  if (/^\d+$/.test(rawValue)) {
    const numericValue = Number(rawValue);
    return Number.isFinite(numericValue) && numericValue > 0
      ? numericValue
      : undefined;
  }
  const parsed = dayjs(rawValue);
  if (!parsed.isValid()) {
    return undefined;
  }
  return boundary === "start"
    ? parsed.startOf("day").valueOf()
    : parsed.endOf("day").valueOf();
}

function invocationTimeRangeParams(formValues: Record<string, any>): {
  endedAt?: number;
  startedAt?: number;
} {
  const range = formValues.createdAtRange;
  if (!Array.isArray(range)) {
    return {};
  }
  const startedAt = resolveRangeBoundary(range[0], "start");
  const endedAt = resolveRangeBoundary(range[1], "end");
  return {
    ...(startedAt ? { startedAt } : {}),
    ...(endedAt ? { endedAt } : {}),
  };
}

function normalizeInvocationFilters(formValues: Record<string, any>) {
  const capability = splitCapabilityMethod(formValues.capabilityKey || "");
  const {
    capabilityKey: _capabilityKey,
    createdAtRange: _createdAtRange,
    ...filters
  } = formValues;
  return {
    capabilityMethod: capability.capabilityMethod,
    capabilityType: capability.capabilityType,
    ...filters,
    ...invocationTimeRangeParams(formValues),
  };
}

function handleDetail(row: Invocation) {
  detailDrawerApi.setData({ record: row });
  detailDrawerApi.open();
}

function handleClean() {
  deleteRange.value = [];
  deleteAllLogs.value = false;
  deleteRangeModalOpen.value = true;
}

async function handleDeleteRangeConfirm() {
  const params = invocationTimeRangeParams({
    createdAtRange: deleteRange.value,
  });
  if (!deleteAllLogs.value && (!params.startedAt || !params.endedAt)) {
    message.warning(
      $t("plugin.linapro-ai-core.invocation.messages.deleteRangeRequired"),
    );
    return;
  }

  deleteRangeSubmitting.value = true;
  try {
    const result = await invocationClean(
      deleteAllLogs.value ? undefined : params,
    );
    message.success(
      $t("plugin.linapro-ai-core.invocation.messages.deleteRangeSuccess", {
        count: result?.deleted ?? 0,
      }),
    );
    deleteRangeModalOpen.value = false;
    await gridApi.query();
  } finally {
    deleteRangeSubmitting.value = false;
  }
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid
      data-testid="ai-invocation-logs-panel"
      :table-title="$t('plugin.linapro-ai-core.invocation.tableTitle')"
    >
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="hasAccessByCodes([accessCodes.clear])"
            danger
            data-testid="ai-invocation-clear"
            type="primary"
            @click="handleClean"
          >
            {{ $t("pages.common.delete") }}
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleDetail(row)">
            {{ $t("pages.common.detail") }}
          </ghost-button>
        </Space>
      </template>
    </Grid>

    <DetailDrawerRef />

    <Modal
      v-model:open="deleteRangeModalOpen"
      :destroy-on-close="true"
      :title="$t('plugin.linapro-ai-core.invocation.messages.deleteRangeTitle')"
    >
      <div>
        <div data-testid="ai-invocation-delete-alert">
          <Alert
            :message="
              $t(
                'plugin.linapro-ai-core.invocation.messages.deleteRangeDescription',
              )
            "
            show-icon
            type="warning"
          />
        </div>
        <div
          data-testid="ai-invocation-delete-all-option"
          style="margin-top: 16px"
        >
          <Checkbox v-model:checked="deleteAllLogs">
            {{
              $t("plugin.linapro-ai-core.invocation.messages.deleteAllLabel")
            }}
          </Checkbox>
          <div class="text-xs text-gray-500" style="margin-top: 4px">
            {{ $t("plugin.linapro-ai-core.invocation.messages.deleteAllHint") }}
          </div>
        </div>
        <div
          data-testid="ai-invocation-delete-range-section"
          style="margin-top: 16px"
        >
          <RangePicker
            v-model:value="deleteRange"
            :disabled="deleteAllLogs"
            :show-time="{
              defaultValue: [dayjs().startOf('day'), dayjs().endOf('day')],
            }"
            class="w-full"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="x"
          />
        </div>
      </div>
      <template #footer>
        <a-button @click="deleteRangeModalOpen = false">
          {{ $t("pages.common.cancel") }}
        </a-button>
        <a-button
          :loading="deleteRangeSubmitting"
          danger
          type="primary"
          @click="handleDeleteRangeConfirm"
        >
          {{ $t("pages.common.confirm") }}
        </a-button>
      </template>
    </Modal>
  </Page>
</template>

<style scoped>
:deep(.ai-invocation-primary-query-field) {
  box-sizing: border-box;
  padding-left: 16px;
}
</style>

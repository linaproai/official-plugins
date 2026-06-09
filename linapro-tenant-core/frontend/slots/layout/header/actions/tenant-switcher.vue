<script lang="ts">
import { pluginCapabilityKeys } from '#/plugins/plugin-capabilities';
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  capabilities: [pluginCapabilityKeys.tenantManagement],
  order: 0,
  slotKey: pluginSlotKeys.layoutHeaderActionsBefore,
};
</script>

<script setup lang="ts">
import { computed, watch } from 'vue';
import { useRouter } from 'vue-router';

import { IconifyIcon } from '@vben/icons';
import { useUserStore } from '@vben/stores';

import { Select, Spin } from 'ant-design-vue';

import { $t } from '#/locales';
import { useTenantStore } from '#/store';

const router = useRouter();
const userStore = useUserStore();
const tenantStore = useTenantStore();

const showTenantSwitcher = computed(() => tenantStore.enabled);

async function handleTenantSwitch(value: unknown) {
  const rawTenantId =
    typeof value === 'object' && value !== null && 'value' in value
      ? (value as { value: unknown }).value
      : value;
  const tenantId = Number(rawTenantId);
  if (
    !Number.isFinite(tenantId) ||
    tenantStore.currentTenant?.id === tenantId
  ) {
    return;
  }
  await tenantStore.switchTenant(tenantId, router);
}

async function handleExitImpersonation() {
  await tenantStore.exitImpersonation(router);
}

watch(
  () => ({
    enabled: tenantStore.enabled,
    isPlatform: tenantStore.isPlatform,
    userId: Number(userStore.userInfo?.userId || 0),
  }),
  ({ enabled, isPlatform, userId }) => {
    if (!enabled) {
      return;
    }
    void tenantStore.ensureTenantOptions({ isPlatform, userId });
  },
  { immediate: true },
);
</script>

<template>
  <div class="tenant-switcher-shell">
    <div
      v-if="tenantStore.isImpersonation"
      class="tenant-impersonation-banner"
      data-testid="impersonation-banner"
    >
      <span
        :title="
          $t('pages.multiTenant.impersonation.banner', {
            tenant: tenantStore.currentTenant?.name || '',
          })
        "
        class="tenant-impersonation-banner__text"
        data-testid="impersonation-banner-text"
      >
        {{
          $t('pages.multiTenant.impersonation.banner', {
            tenant: tenantStore.currentTenant?.name || '',
          })
        }}
      </span>
      <a-button
        danger
        ghost
        size="small"
        class="tenant-impersonation-banner__exit"
        data-testid="impersonation-exit"
        @click="handleExitImpersonation"
      >
        {{ $t('pages.multiTenant.impersonation.exit') }}
      </a-button>
    </div>
    <div v-if="showTenantSwitcher" data-testid="tenant-switcher">
      <Select
        :value="tenantStore.currentTenant?.id"
        :disabled="tenantStore.isImpersonation"
        :field-names="{ label: 'name', value: 'id' }"
        :filter-option="
          (input, option) =>
            String(option?.name || '')
              .toLowerCase()
              .includes(input.toLowerCase()) ||
            String(option?.code || '')
              .toLowerCase()
              .includes(input.toLowerCase())
        "
        :not-found-content="$t('pages.multiTenant.empty.tenants')"
        :options="tenantStore.tenants"
        :placeholder="$t('pages.multiTenant.switcher.placeholder')"
        class="tenant-switcher__select"
        data-testid="tenant-switcher-select"
        show-search
        @select="handleTenantSwitch"
      >
        <template #suffixIcon>
          <Spin
            v-if="tenantStore.switching || tenantStore.loadingTenants"
            size="small"
            spinning
          />
          <IconifyIcon
            v-else
            class="tenant-switcher__icon"
            icon="lucide:building-2"
          />
        </template>
        <template #option="{ name, code }">
          <div class="tenant-switcher-option">
            <span class="tenant-switcher-option__name">{{ name }}</span>
            <span class="tenant-switcher-option__code">{{ code }}</span>
          </div>
        </template>
      </Select>
    </div>
  </div>
</template>

<style scoped>
.tenant-switcher-shell {
  display: none;
  align-items: center;
  margin-right: 8px;
}

@media (min-width: 768px) {
  .tenant-switcher-shell {
    display: flex;
  }
}

.tenant-impersonation-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  gap: 6px;
  height: 32px;
  max-width: 220px;
  padding: 0 10px;
  margin-right: 8px;
  color: var(--tenant-impersonation-banner-color, rgb(185 28 28));
  font-size: 12px;
  font-weight: 500;
  background: var(--tenant-impersonation-banner-bg, rgb(254 242 242));
  border: 1px solid
    var(--tenant-impersonation-banner-border, rgb(252 165 165));
  border-radius: 4px;
}

@media (min-width: 1280px) {
  .tenant-impersonation-banner {
    max-width: 240px;
  }
}

:global(.dark) .tenant-impersonation-banner {
  --tenant-impersonation-banner-color: rgb(254 202 202);
  --tenant-impersonation-banner-bg: rgb(239 68 68 / 15%);
  --tenant-impersonation-banner-border: rgb(239 68 68 / 60%);
}

.tenant-impersonation-banner__text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tenant-impersonation-banner__exit {
  flex-shrink: 0;
  white-space: nowrap;
}

.tenant-switcher__select {
  width: 240px;
  min-width: 240px;
}

.tenant-switcher__icon {
  width: 16px;
  height: 16px;
}

.tenant-switcher-option {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.tenant-switcher-option__name,
.tenant-switcher-option__code {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tenant-switcher-option__code {
  color: hsl(var(--muted-foreground));
  font-size: 12px;
}
</style>

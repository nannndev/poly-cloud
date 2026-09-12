<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

defineProps<{
  collapsed?: boolean
}>()

const colorMode = useColorMode()
const config = useRuntimeConfig()

// v1 berjalan single-user tanpa login (docs/06), jadi tak ada identitas pengguna
// untuk ditampilkan — yang relevan adalah instans backend yang sedang dipakai.
const apiHost = computed(() => {
  try {
    return new URL(config.public.apiBase).host
  } catch {
    return config.public.apiBase
  }
})

const items = computed<DropdownMenuItem[][]>(() => [
  [
    {
      type: 'label',
      label: 'Poly Cloud',
      icon: 'i-lucide-hard-drive'
    }
  ],
  [
    {
      label: 'Manage Cloud Accounts',
      icon: 'i-lucide-cloud-cog',
      to: '/accounts'
    },
    {
      label: 'Storage & Quota Analytics',
      icon: 'i-lucide-pie-chart',
      to: '/quota'
    },
    {
      label: 'Routing & Settings',
      icon: 'i-lucide-sliders-horizontal',
      to: '/settings'
    }
  ],
  [
    {
      label: colorMode.value === 'dark' ? 'Light Mode' : 'Dark Mode',
      icon: colorMode.value === 'dark' ? 'i-lucide-sun' : 'i-lucide-moon',
      onSelect() {
        colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'
      }
    }
  ],
  [
    {
      label: 'API Health Check',
      icon: 'i-lucide-external-link',
      to: config.public.apiBase.replace(/\/api\/v1$/, '') + '/healthz',
      target: '_blank'
    }
  ]
])
</script>

<template>
  <UDropdownMenu
    :items="items"
    :content="{ align: 'center', collisionPadding: 12 }"
    :ui="{ content: collapsed ? 'w-56' : 'w-(--reka-dropdown-menu-trigger-width)' }"
  >
    <div
      class="group flex w-full items-center gap-2.5 rounded-xl border border-transparent p-2 text-left transition-all duration-200 hover:border-default/80 hover:bg-elevated/50 cursor-pointer"
      :class="[collapsed ? 'justify-center p-1.5' : '']"
    >
      <!-- Penanda instans, bukan identitas pengguna (v1 belum punya login) -->
      <div class="relative shrink-0">
        <div class="size-8 rounded-full border border-default bg-elevated flex items-center justify-center">
          <UIcon name="i-lucide-hard-drive" class="size-4 text-emerald-400" />
        </div>
        <span class="absolute bottom-0 right-0 size-2 rounded-full bg-emerald-500 ring-2 ring-card" />
      </div>

      <div v-if="!collapsed" class="min-w-0 flex-1 leading-tight space-y-0.5">
        <p class="font-bold text-xs text-highlight truncate">Poly Cloud</p>
        <p class="text-[10px] text-muted truncate" :title="config.public.apiBase">
          {{ apiHost }}
        </p>
      </div>

      <UIcon
        v-if="!collapsed"
        name="i-lucide-chevrons-up-down"
        class="size-3.5 text-muted shrink-0 ml-auto transition-transform duration-200 group-hover:text-highlight"
      />
    </div>
  </UDropdownMenu>
</template>

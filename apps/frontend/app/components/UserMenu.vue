<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

defineProps<{
  collapsed?: boolean
}>()

const colorMode = useColorMode()
const config = useRuntimeConfig()
const { user, logout } = useAuth()

const apiHost = computed(() => {
  try {
    return new URL(config.public.apiBase).host
  } catch {
    return config.public.apiBase
  }
})

function toggleTheme() {
  const isDark = colorMode.value === 'dark' || (import.meta.client && document.documentElement.classList.contains('dark'))
  const next = isDark ? 'light' : 'dark'
  colorMode.preference = next
  if (import.meta.client) {
    if (next === 'light') {
      document.documentElement.classList.remove('dark')
      document.documentElement.classList.add('light')
    } else {
      document.documentElement.classList.remove('light')
      document.documentElement.classList.add('dark')
    }
  }
}

const items = computed<DropdownMenuItem[][]>(() => [
  [
    {
      type: 'label',
      label: user.value?.email || 'admin@polycloud.local',
      icon: 'i-lucide-user-check'
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
        toggleTheme()
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
  ],
  [
    {
      label: 'Sign Out',
      icon: 'i-lucide-log-out',
      async onSelect() {
        await logout()
      }
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
      <div class="relative shrink-0">
        <div class="flex size-8 items-center justify-center rounded-full border border-brand-400/25 bg-[#1E2430]">
          <PolyMark class="size-[18px]" />
        </div>
        <span class="absolute bottom-0 right-0 size-2 rounded-full bg-primary-500 ring-2 ring-white dark:ring-card" />
      </div>

      <div v-if="!collapsed" class="min-w-0 flex-1 leading-tight space-y-0.5">
        <p class="font-bold text-xs text-slate-900 dark:text-highlight truncate">
          {{ user?.email || 'admin@polycloud.local' }}
        </p>
        <p class="text-[10px] text-slate-500 dark:text-muted truncate" :title="config.public.apiBase">
          Admin • {{ apiHost }}
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

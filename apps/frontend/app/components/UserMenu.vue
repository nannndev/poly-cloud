<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

defineProps<{
  collapsed?: boolean
}>()

const colorMode = useColorMode()
const accountsStore = useAccountsStore()

const items = computed<DropdownMenuItem[][]>(() => [
  [
    {
      type: 'label',
      label: 'Admin Workspace',
      avatar: {
        src: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&auto=format&fit=crop&q=80',
        alt: 'Yubidev Admin'
      }
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
      label: 'REST API Health Docs',
      icon: 'i-lucide-external-link',
      to: 'http://localhost:8080/healthz',
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
      <!-- User Avatar -->
      <div class="relative shrink-0">
        <div
          class="size-8 rounded-full border border-default bg-elevated bg-cover bg-center"
          style="background-image: url('https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&auto=format&fit=crop&q=80');"
        />
        <span class="absolute bottom-0 right-0 size-2 rounded-full bg-emerald-500 ring-2 ring-card" />
      </div>

      <!-- Info -->
      <div v-if="!collapsed" class="min-w-0 flex-1 leading-tight space-y-0.5">
        <div class="flex items-center gap-1.5">
          <p class="font-bold text-xs text-highlight truncate">
            Yubidev Admin
          </p>
        </div>
        <p class="text-[10px] text-muted truncate">
          {{ accountsStore.activeAccounts.length }} Active Cloud Drives
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

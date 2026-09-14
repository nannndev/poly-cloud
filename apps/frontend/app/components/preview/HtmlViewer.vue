<script setup lang="ts">
import CodeViewer from './CodeViewer.vue'

const props = withDefaults(defineProps<{
  htmlContent: string
  filename?: string
  fileSize?: number
}>(), {
  filename: '',
  fileSize: 0
})

const mode = ref<'preview' | 'source'>('preview')
const viewport = ref<'desktop' | 'tablet' | 'mobile'>('desktop')
const iframeKey = ref(0)

const viewportWidthClass = computed(() => {
  if (viewport.value === 'mobile') return 'w-[375px] max-w-full'
  if (viewport.value === 'tablet') return 'w-[768px] max-w-full'
  return 'w-full'
})

function reloadIframe() {
  iframeKey.value++
}

function openInNewTab() {
  const blob = new Blob([props.htmlContent], { type: 'text/html' })
  const url = URL.createObjectURL(blob)
  window.open(url, '_blank')
}
</script>

<template>
  <div class="flex flex-col h-full min-h-0 rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-slate-50 dark:bg-[#090d15] overflow-hidden">
    <!-- Top Toolbar -->
    <div class="flex shrink-0 flex-wrap items-center justify-between gap-3 px-4 py-2 border-b border-slate-200/80 dark:border-white/[0.06] bg-white/80 dark:bg-[#121723]/90 text-xs backdrop-blur-sm">
      <!-- Left: Mode Switcher -->
      <div class="flex items-center gap-2">
        <div class="flex items-center rounded-xl border border-slate-300/80 dark:border-white/[0.08] bg-slate-100 dark:bg-[#1a202f] p-0.5 shadow-2xs">
          <button
            type="button"
            class="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold transition-all cursor-pointer"
            :class="[
              mode === 'preview'
                ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                : 'text-slate-600 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-white'
            ]"
            @click="mode = 'preview'"
          >
            <UIcon name="i-lucide-globe" class="size-3.5" />
            <span>Live Preview</span>
          </button>

          <button
            type="button"
            class="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold transition-all cursor-pointer"
            :class="[
              mode === 'source'
                ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                : 'text-slate-600 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-white'
            ]"
            @click="mode = 'source'"
          >
            <UIcon name="i-lucide-code-2" class="size-3.5" />
            <span>Source Code</span>
          </button>
        </div>

        <span v-if="filename" class="font-mono text-slate-800 dark:text-zinc-200 truncate font-medium text-[11px] hidden sm:inline">
          {{ filename }}
        </span>
      </div>

      <!-- Right: Responsive Viewport Switcher & External Tab -->
      <div class="flex items-center gap-2">
        <!-- Viewport Switcher (Only in preview mode) -->
        <div v-if="mode === 'preview'" class="flex items-center rounded-xl border border-slate-300/80 dark:border-white/[0.08] bg-slate-100 dark:bg-[#1a202f] p-0.5 shadow-2xs">
          <button
            type="button"
            title="Desktop view (100%)"
            class="p-1.5 rounded-lg transition-colors cursor-pointer"
            :class="[
              viewport === 'desktop'
                ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                : 'text-slate-500 hover:text-slate-900 dark:text-zinc-400 dark:hover:text-white'
            ]"
            @click="viewport = 'desktop'"
          >
            <UIcon name="i-lucide-monitor" class="size-3.5" />
          </button>

          <button
            type="button"
            title="Tablet view (768px)"
            class="p-1.5 rounded-lg transition-colors cursor-pointer"
            :class="[
              viewport === 'tablet'
                ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                : 'text-slate-500 hover:text-slate-900 dark:text-zinc-400 dark:hover:text-white'
            ]"
            @click="viewport = 'tablet'"
          >
            <UIcon name="i-lucide-tablet" class="size-3.5" />
          </button>

          <button
            type="button"
            title="Mobile view (375px)"
            class="p-1.5 rounded-lg transition-colors cursor-pointer"
            :class="[
              viewport === 'mobile'
                ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                : 'text-slate-500 hover:text-slate-900 dark:text-zinc-400 dark:hover:text-white'
            ]"
            @click="viewport = 'mobile'"
          >
            <UIcon name="i-lucide-smartphone" class="size-3.5" />
          </button>
        </div>

        <!-- Reload Iframe -->
        <button
          v-if="mode === 'preview'"
          type="button"
          title="Reload preview"
          class="flex size-7 items-center justify-center rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-200/60 dark:text-zinc-400 dark:hover:text-white dark:hover:bg-white/[0.08] transition-colors cursor-pointer"
          @click="reloadIframe"
        >
          <UIcon name="i-lucide-refresh-cw" class="size-3.5" />
        </button>

        <!-- Open in New Tab -->
        <button
          type="button"
          title="Open in new window"
          class="flex size-7 items-center justify-center rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-200/60 dark:text-zinc-400 dark:hover:text-white dark:hover:bg-white/[0.08] transition-colors cursor-pointer"
          @click="openInNewTab"
        >
          <UIcon name="i-lucide-external-link" class="size-3.5" />
        </button>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 min-h-0 overflow-hidden flex flex-col">
      <!-- LIVE HTML PREVIEW -->
      <div
        v-if="mode === 'preview'"
        class="flex-1 min-h-0 overflow-auto bg-slate-200/50 dark:bg-[#070a10] p-4 flex items-center justify-center"
      >
        <div
          class="h-full bg-white shadow-2xl rounded-xl overflow-hidden border border-slate-300 dark:border-white/10 transition-all duration-300 flex flex-col"
          :class="viewportWidthClass"
        >
          <iframe
            :key="iframeKey"
            :srcdoc="htmlContent"
            title="HTML Live Preview"
            sandbox="allow-scripts allow-same-origin"
            class="w-full h-full border-none bg-white"
          />
        </div>
      </div>

      <!-- SOURCE CODE VIEW -->
      <div v-else class="flex-1 min-h-0">
        <CodeViewer
          :code="htmlContent"
          language="html"
          :filename="filename"
          :file-size="fileSize"
          :show-toolbar="false"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import CodeViewer from './CodeViewer.vue'
import JsonTreeNode from './JsonTreeNode.vue'

const props = withDefaults(defineProps<{
  rawJson: string
  filename?: string
  fileSize?: number
}>(), {
  filename: '',
  fileSize: 0
})

const mode = ref<'tree' | 'raw'>('tree')
const searchQuery = ref('')
const isCopied = ref(false)
const treeKey = ref(0) // increment to force tree expand/collapse reset

interface ParseResult {
  valid: boolean
  data: any
  error?: string
  formatted: string
  summary: string
}

const parsedJson = computed<ParseResult>(() => {
  if (!props.rawJson) {
    return { valid: false, data: null, error: 'Empty JSON content', formatted: '', summary: '0 items' }
  }
  try {
    const data = JSON.parse(props.rawJson)
    const formatted = JSON.stringify(data, null, 2)
    let summary = ''
    if (Array.isArray(data)) {
      summary = `${data.length} ${data.length === 1 ? 'item' : 'items'}`
    } else if (typeof data === 'object' && data !== null) {
      const keys = Object.keys(data).length
      summary = `${keys} ${keys === 1 ? 'key' : 'keys'}`
    } else {
      summary = typeof data
    }
    return { valid: true, data, formatted, summary }
  } catch (err: any) {
    return {
      valid: false,
      data: null,
      error: err?.message || 'Invalid JSON syntax',
      formatted: props.rawJson,
      summary: 'Syntax Error'
    }
  }
})

function copyFormatted() {
  navigator.clipboard?.writeText(parsedJson.value.formatted || props.rawJson)
  isCopied.value = true
  setTimeout(() => {
    isCopied.value = false
  }, 2000)
}

function resetTree() {
  treeKey.value++
}
</script>

<template>
  <div class="flex flex-col h-full min-h-0 rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-slate-50 dark:bg-[#090d15] overflow-hidden">
    <!-- Top Toolbar -->
    <div class="flex shrink-0 flex-wrap items-center justify-between gap-3 px-4 py-2 border-b border-slate-200/80 dark:border-white/[0.06] bg-white/80 dark:bg-[#121723]/90 text-xs backdrop-blur-sm">
      <!-- Left: Status & Meta -->
      <div class="flex items-center gap-2.5 min-w-0">
        <!-- Status Pill -->
        <span
          class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-lg font-mono text-[10px] font-bold uppercase tracking-wider border"
          :class="[
            parsedJson.valid
              ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20'
              : 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20'
          ]"
        >
          <UIcon :name="parsedJson.valid ? 'i-lucide-check-circle-2' : 'i-lucide-alert-triangle'" class="size-3" />
          {{ parsedJson.valid ? 'Valid JSON' : 'Invalid JSON' }}
        </span>

        <span v-if="parsedJson.valid" class="font-mono text-slate-500 dark:text-zinc-400 text-[11px]">
          {{ parsedJson.summary }}
        </span>

        <span v-if="filename" class="font-mono text-slate-800 dark:text-zinc-200 truncate font-medium text-[11px] hidden sm:inline">
          • {{ filename }}
        </span>
      </div>

      <!-- Right: View Switcher & Actions -->
      <div class="flex items-center gap-2 shrink-0">
        <!-- Search filter for tree -->
        <div v-if="mode === 'tree' && parsedJson.valid" class="flex items-center gap-1 bg-slate-100 dark:bg-[#1c2231] px-2 py-1 rounded-xl border border-slate-300/80 dark:border-white/10 text-xs">
          <UIcon name="i-lucide-search" class="size-3 text-slate-500 dark:text-zinc-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Filter key/value..."
            class="bg-transparent border-none outline-none text-[11px] text-slate-800 dark:text-white w-24 sm:w-32 font-mono"
          >
          <button
            v-if="searchQuery"
            type="button"
            class="text-slate-400 hover:text-slate-700 dark:hover:text-white cursor-pointer"
            @click="searchQuery = ''"
          >
            <UIcon name="i-lucide-x" class="size-3" />
          </button>
        </div>

        <!-- Mode Switcher -->
        <div class="flex items-center rounded-xl border border-slate-300/80 dark:border-white/[0.08] bg-slate-100 dark:bg-[#1a202f] p-0.5 shadow-2xs">
          <button
            type="button"
            class="flex items-center gap-1 px-2.5 py-1 rounded-lg text-[11px] font-semibold transition-all cursor-pointer"
            :class="[
              mode === 'tree'
                ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                : 'text-slate-600 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-white'
            ]"
            :disabled="!parsedJson.valid"
            @click="mode = 'tree'"
          >
            <UIcon name="i-lucide-network" class="size-3.5" />
            <span>Tree</span>
          </button>

          <button
            type="button"
            class="flex items-center gap-1 px-2.5 py-1 rounded-lg text-[11px] font-semibold transition-all cursor-pointer"
            :class="[
              mode === 'raw'
                ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                : 'text-slate-600 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-white'
            ]"
            @click="mode = 'raw'"
          >
            <UIcon name="i-lucide-code-2" class="size-3.5" />
            <span>Formatted</span>
          </button>
        </div>

        <!-- Copy Action -->
        <button
          type="button"
          class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-semibold transition-all cursor-pointer shadow-2xs"
          :class="[
            isCopied
              ? 'bg-emerald-500 text-white'
              : 'bg-slate-200/80 hover:bg-slate-300/80 text-slate-700 hover:text-slate-900 dark:bg-[#202738] dark:hover:bg-[#283248] dark:text-zinc-200 dark:hover:text-white border border-slate-300/80 dark:border-white/10'
          ]"
          @click="copyFormatted"
        >
          <UIcon :name="isCopied ? 'i-lucide-check' : 'i-lucide-copy'" class="size-3.5" />
          <span>{{ isCopied ? 'Copied' : 'Copy' }}</span>
        </button>
      </div>
    </div>

    <!-- Error Banner if invalid -->
    <div
      v-if="!parsedJson.valid"
      class="px-4 py-2.5 bg-rose-500/10 border-b border-rose-500/20 text-rose-600 dark:text-rose-400 text-xs flex items-center gap-2 shrink-0 font-mono"
    >
      <UIcon name="i-lucide-circle-alert" class="size-4 shrink-0" />
      <span>JSON Syntax Error: {{ parsedJson.error }}</span>
    </div>

    <!-- Content Area: Tree Mode vs Raw Code Mode -->
    <div class="flex-1 min-h-0 overflow-hidden flex flex-col">
      <!-- TREE VIEW -->
      <div
        v-if="mode === 'tree' && parsedJson.valid"
        class="flex-1 min-h-0 overflow-auto p-4 select-text"
      >
        <JsonTreeNode
          :key="treeKey"
          :data="parsedJson.data"
          :depth="0"
          :search="searchQuery"
        />
      </div>

      <!-- FORMATTED RAW VIEW -->
      <div v-else class="flex-1 min-h-0">
        <CodeViewer
          :code="parsedJson.formatted"
          language="json"
          :filename="filename"
          :file-size="fileSize"
          :show-toolbar="false"
        />
      </div>
    </div>
  </div>
</template>

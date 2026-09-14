<script setup lang="ts">
import CodeViewer from './CodeViewer.vue'

const props = withDefaults(defineProps<{
  markdown: string
  filename?: string
  fileSize?: number
}>(), {
  filename: '',
  fileSize: 0
})

const mode = ref<'rendered' | 'source'>('rendered')

// Simple safe markdown to HTML parser without heavy external libs
const renderedHtml = computed(() => {
  if (!props.markdown) return ''
  let text = props.markdown

  // Escape basic HTML
  text = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

  // Code blocks ```code```
  text = text.replace(/```([a-z0-9_-]*)\n([\s\S]*?)```/g, (_, _lang, code) => {
    return `<pre class="my-3 p-3.5 rounded-xl bg-slate-100 dark:bg-[#151c2c] border border-slate-200 dark:border-white/10 font-mono text-xs overflow-x-auto text-slate-800 dark:text-zinc-200"><code>${code.trim()}</code></pre>`
  })

  // Inline code `code`
  text = text.replace(/`([^`]+)`/g, '<code class="px-1.5 py-0.5 rounded bg-slate-200/80 dark:bg-zinc-800 font-mono text-[11px] text-primary-600 dark:text-primary-400 font-semibold">$1</code>')

  // Headers
  text = text.replace(/^### (.*$)/gim, '<h3 class="text-base font-bold text-slate-900 dark:text-white mt-4 mb-2">$1</h3>')
  text = text.replace(/^## (.*$)/gim, '<h2 class="text-lg font-bold text-slate-900 dark:text-white mt-5 mb-2.5 pb-1 border-b border-slate-200 dark:border-white/10">$1</h2>')
  text = text.replace(/^# (.*$)/gim, '<h1 class="text-xl font-black text-slate-900 dark:text-white mt-6 mb-3 pb-1.5 border-b border-slate-200 dark:border-white/10">$1</h1>')

  // Blockquotes
  text = text.replace(/^\> (.*$)/gim, '<blockquote class="my-2 pl-3.5 border-l-2 border-primary-500 text-slate-600 dark:text-zinc-400 italic text-xs">$1</blockquote>')

  // Bold & Italic
  text = text.replace(/\*\*\*(.*?)\*\*\*/g, '<strong><em>$1</em></strong>')
  text = text.replace(/\*\*(.*?)\*\*/g, '<strong class="font-bold text-slate-900 dark:text-white">$1</strong>')
  text = text.replace(/\*(.*?)\*/g, '<em class="italic">$1</em>')

  // Links
  text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer" class="text-primary-600 dark:text-primary-400 underline underline-offset-2 hover:text-primary-500">$1</a>')

  // Unordered list items
  text = text.replace(/^\s*[-*]\s+(.*$)/gim, '<li class="ml-4 list-disc text-slate-700 dark:text-zinc-300">$1</li>')

  // Paragraphs & Line Breaks
  const lines = text.split('\n')
  const processed: string[] = []
  let inList = false

  for (const line of lines) {
    if (line.startsWith('<li')) {
      if (!inList) {
        processed.push('<ul class="my-2 space-y-1">')
        inList = true
      }
      processed.push(line)
    } else {
      if (inList) {
        processed.push('</ul>')
        inList = false
      }
      if (line.trim().length === 0) {
        processed.push('<div class="h-2"></div>')
      } else if (!line.startsWith('<h') && !line.startsWith('<pre') && !line.startsWith('<block')) {
        processed.push(`<p class="leading-relaxed text-slate-700 dark:text-zinc-300 text-xs">${line}</p>`)
      } else {
        processed.push(line)
      }
    }
  }
  if (inList) processed.push('</ul>')

  return processed.join('\n')
})
</script>

<template>
  <div class="flex flex-col h-full min-h-0 rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-slate-50 dark:bg-[#090d15] overflow-hidden">
    <!-- Top Toolbar -->
    <div class="flex shrink-0 items-center justify-between gap-3 px-4 py-2 border-b border-slate-200/80 dark:border-white/[0.06] bg-white/80 dark:bg-[#121723]/90 text-xs backdrop-blur-sm">
      <!-- Left: Mode Switcher -->
      <div class="flex items-center gap-2">
        <div class="flex items-center rounded-xl border border-slate-300/80 dark:border-white/[0.08] bg-slate-100 dark:bg-[#1a202f] p-0.5 shadow-2xs">
          <button
            type="button"
            class="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold transition-all cursor-pointer"
            :class="[
              mode === 'rendered'
                ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                : 'text-slate-600 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-white'
            ]"
            @click="mode = 'rendered'"
          >
            <UIcon name="i-lucide-book-open" class="size-3.5" />
            <span>Preview</span>
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
            <span>Markdown Source</span>
          </button>
        </div>

        <span v-if="filename" class="font-mono text-slate-800 dark:text-zinc-200 truncate font-medium text-[11px] hidden sm:inline">
          {{ filename }}
        </span>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 min-h-0 overflow-hidden flex flex-col">
      <!-- RENDERED DOCUMENT -->
      <div
        v-if="mode === 'rendered'"
        class="flex-1 min-h-0 overflow-auto p-6 bg-white dark:bg-[#0a0d15] selection:bg-primary-500/25"
      >
        <article class="max-w-3xl mx-auto" v-html="renderedHtml" />
      </div>

      <!-- MARKDOWN SOURCE -->
      <div v-else class="flex-1 min-h-0">
        <CodeViewer
          :code="markdown"
          language="markdown"
          :filename="filename"
          :file-size="fileSize"
          :show-toolbar="false"
        />
      </div>
    </div>
  </div>
</template>

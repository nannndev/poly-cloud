<script setup lang="ts">
interface Props {
  nodeKey?: string | number
  data: any
  depth?: number
  isLast?: boolean
  search?: string
}

const props = withDefaults(defineProps<Props>(), {
  nodeKey: undefined,
  depth: 0,
  isLast: true,
  search: ''
})

const isOpen = ref(props.depth < 2)

const dataType = computed(() => {
  if (props.data === null) return 'null'
  if (Array.isArray(props.data)) return 'array'
  return typeof props.data
})

const isComplex = computed(() => dataType.value === 'object' || dataType.value === 'array')

const childKeys = computed(() => {
  if (dataType.value === 'array') {
    return (props.data as any[]).map((_, idx) => idx)
  }
  if (dataType.value === 'object' && props.data !== null) {
    return Object.keys(props.data)
  }
  return []
})

const itemSummary = computed(() => {
  if (dataType.value === 'array') {
    const len = (props.data as any[]).length
    return len === 1 ? '1 item' : `${len} items`
  }
  if (dataType.value === 'object' && props.data !== null) {
    const len = Object.keys(props.data).length
    return len === 1 ? '1 key' : `${len} keys`
  }
  return ''
})

const isSearchMatch = computed(() => {
  if (!props.search.trim()) return false
  const q = props.search.toLowerCase()
  const keyMatch = props.nodeKey !== undefined && String(props.nodeKey).toLowerCase().includes(q)
  const valMatch = !isComplex.value && String(props.data).toLowerCase().includes(q)
  return keyMatch || valMatch
})

watch(() => props.search, (val) => {
  if (val && isComplex.value) {
    isOpen.value = true
  }
})
</script>

<template>
  <div class="font-mono text-xs leading-relaxed select-text" :class="{ 'bg-amber-500/10 rounded': isSearchMatch }">
    <!-- Complex Node Header (Object or Array) -->
    <div
      v-if="isComplex"
      class="inline-flex items-center gap-1.5 py-0.5 rounded px-1 hover:bg-slate-200/60 dark:hover:bg-white/[0.04] transition-colors cursor-pointer"
      @click="isOpen = !isOpen"
    >
      <!-- Expand/Collapse Chevron -->
      <button
        type="button"
        class="size-4 flex items-center justify-center text-slate-400 dark:text-zinc-500 hover:text-slate-900 dark:hover:text-white shrink-0"
      >
        <UIcon
          :name="isOpen ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
          class="size-3.5 transition-transform"
        />
      </button>

      <!-- Key Name -->
      <span v-if="nodeKey !== undefined" class="text-sky-600 dark:text-sky-400 font-semibold">
        "{{ nodeKey }}":
      </span>

      <!-- Opening Bracket / Summary -->
      <span class="text-slate-600 dark:text-zinc-400 font-bold">
        {{ dataType === 'array' ? '[' : '{' }}
      </span>

      <span v-if="!isOpen" class="text-[11px] text-slate-500 dark:text-zinc-500 px-1.5 py-0.5 rounded bg-slate-200/80 dark:bg-zinc-800/80 font-normal">
        {{ itemSummary }}
      </span>

      <span v-if="!isOpen" class="text-slate-600 dark:text-zinc-400 font-bold">
        {{ dataType === 'array' ? ']' : '}' }}{{ isLast ? '' : ',' }}
      </span>
    </div>

    <!-- Primitive Node (String, Number, Boolean, Null) -->
    <div
      v-else
      class="inline-flex items-center gap-1 py-0.5 px-1 rounded hover:bg-slate-200/60 dark:hover:bg-white/[0.04] transition-colors"
    >
      <span class="size-4 shrink-0" />
      <span v-if="nodeKey !== undefined" class="text-sky-600 dark:text-sky-400 font-semibold">
        "{{ nodeKey }}":
      </span>

      <!-- Value Rendering -->
      <span v-if="dataType === 'string'" class="text-emerald-600 dark:text-emerald-400 break-all">
        "{{ data }}"
      </span>
      <span v-else-if="dataType === 'number'" class="text-amber-600 dark:text-amber-400 font-semibold">
        {{ data }}
      </span>
      <span v-else-if="dataType === 'boolean'" class="text-violet-600 dark:text-violet-400 font-bold">
        {{ data }}
      </span>
      <span v-else-if="dataType === 'null'" class="text-rose-500 font-bold italic">
        null
      </span>
      <span v-else class="text-slate-600 dark:text-zinc-400">
        {{ String(data) }}
      </span>

      <span v-if="!isLast" class="text-slate-400 dark:text-zinc-600">,</span>
    </div>

    <!-- Nested Children -->
    <div v-if="isComplex && isOpen" class="pl-5 border-l border-slate-200 dark:border-white/[0.06] ml-2">
      <JsonTreeNode
        v-for="(childKey, idx) in childKeys"
        :key="childKey"
        :node-key="dataType === 'array' ? undefined : childKey"
        :data="data[childKey]"
        :depth="depth + 1"
        :is-last="idx === childKeys.length - 1"
        :search="search"
      />
      <!-- Closing Bracket -->
      <div class="py-0.5 text-slate-600 dark:text-zinc-400 font-bold">
        {{ dataType === 'array' ? ']' : '}' }}{{ isLast ? '' : ',' }}
      </div>
    </div>
  </div>
</template>

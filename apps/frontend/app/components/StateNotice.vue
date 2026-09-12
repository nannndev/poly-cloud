<script setup lang="ts">
/**
 * Panel status bersama untuk daftar: memuat, gagal, atau kosong. Tanpa ini,
 * backend yang mati tampak sama persis dengan folder yang memang kosong.
 */
withDefaults(defineProps<{
  variant?: 'loading' | 'error' | 'empty'
  title?: string
  description?: string
  icon?: string
  retryLabel?: string
}>(), {
  variant: 'empty',
  icon: 'i-lucide-folder-open',
  retryLabel: 'Try again'
})

defineEmits<{ retry: [] }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-3 py-16 text-center">
    <div
      class="p-4 rounded-3xl border"
      :class="variant === 'error'
        ? 'bg-rose-500/[0.06] border-rose-500/25 text-rose-400'
        : 'bg-[#121215] border-white/[0.08] text-zinc-500'"
    >
      <UIcon
        :name="variant === 'loading' ? 'i-lucide-loader-2' : variant === 'error' ? 'i-lucide-circle-alert' : icon"
        class="size-10"
        :class="variant === 'loading' ? 'animate-spin text-emerald-400' : ''"
      />
    </div>

    <div>
      <h4 class="font-bold text-sm" :class="variant === 'error' ? 'text-rose-300' : 'text-zinc-200'">
        {{ title }}
      </h4>
      <p v-if="description" class="text-xs text-zinc-400 mt-0.5 max-w-sm">
        {{ description }}
      </p>
    </div>

    <div v-if="variant !== 'loading'" class="flex items-center gap-2 mt-1">
      <UButton
        v-if="variant === 'error'"
        :label="retryLabel"
        icon="i-lucide-refresh-cw"
        color="error"
        variant="subtle"
        size="xs"
        class="rounded-xl font-semibold"
        @click="$emit('retry')"
      />
      <slot name="actions" />
    </div>
  </div>
</template>

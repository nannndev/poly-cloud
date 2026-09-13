<script setup lang="ts">
/**
 * Lapisan yang muncul saat file diseret dari sistem berkas ke jendela.
 *
 * Tujuan unggahan disebut secara eksplisit: unggahan mendarat di folder yang
 * sedang dibuka, bukan di root, dan itu satu-satunya kesempatan memberi tahu
 * sebelum berkasnya terlanjur terkirim.
 */
defineProps<{
  show: boolean
  /** Path folder yang sedang dibuka, mis. "/Documents". */
  destination: string
}>()
</script>

<template>
  <Transition
    enter-active-class="transition-opacity duration-150"
    leave-active-class="transition-opacity duration-150"
    enter-from-class="opacity-0"
    leave-to-class="opacity-0"
  >
    <!-- pointer-events-none: lapisan ini hanya penanda visual. Menangkap
         pointer akan membuat elemen di bawahnya kehilangan event drop. -->
    <div
      v-if="show"
      class="pointer-events-none fixed inset-0 z-50 flex items-center justify-center bg-[#0b0e14]/85 backdrop-blur-sm"
      aria-hidden="true"
    >
      <div class="mx-4 flex max-w-md flex-col items-center gap-5 rounded-3xl border-2 border-dashed border-primary-400/50 bg-[#141925] px-10 py-12 text-center shadow-2xl">
        <div class="flex size-16 items-center justify-center rounded-2xl border border-primary-500/25 bg-primary-500/15 text-primary-400">
          <UIcon name="i-lucide-upload-cloud" class="size-8" />
        </div>

        <div>
          <p class="text-lg font-bold text-white">Drop to upload</p>
          <p class="mt-1.5 text-sm text-zinc-400">
            Files will be added to
            <strong class="font-mono font-semibold text-primary-400">{{ destination }}</strong>
          </p>
        </div>

        <p class="max-w-xs text-[11px] leading-relaxed text-zinc-500">
          The destination account is chosen automatically — whichever has the
          most free space.
        </p>
      </div>
    </div>
  </Transition>
</template>

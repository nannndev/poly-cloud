<script setup lang="ts">
definePageMeta({
  layout: false
})

const { login } = useAuth()
const router = useRouter()
const toast = useToast()
const colorMode = useColorMode()

const email = ref('admin@polycloud.local')
const password = ref('admin123')
const showPassword = ref(false)
const isLoading = ref(false)
const errorMessage = ref('')

function toggleColorMode() {
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

async function handleSubmit() {
  if (!email.value || !password.value) {
    errorMessage.value = 'Please enter both email and password'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    await login(email.value.trim(), password.value)
    toast.add({
      title: 'Welcome to Poly Cloud',
      description: 'Signed in successfully',
      color: 'success'
    })
    await router.push('/')
  } catch (err: any) {
    errorMessage.value = err?.message || 'Invalid email or password'
  } finally {
    isLoading.value = false
  }
}

function fillDefaultCredentials() {
  email.value = 'admin@polycloud.local'
  password.value = 'admin123'
  errorMessage.value = ''
}
</script>

<template>
  <div class="relative flex min-h-screen w-full items-center justify-center overflow-hidden bg-slate-50 dark:bg-[#07090e] px-4 py-12 text-slate-900 dark:text-slate-100 transition-colors duration-200">
    <!-- Floating Quick Theme Switcher on Top Right -->
    <div class="absolute top-4 right-4 z-20">
      <button
        type="button"
        :title="colorMode.value === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'"
        class="flex size-9 items-center justify-center rounded-xl border border-slate-200 dark:border-white/10 bg-white/80 dark:bg-white/[0.05] text-slate-600 dark:text-zinc-300 backdrop-blur-md shadow-xs hover:bg-slate-100 dark:hover:bg-white/[0.1] transition-all cursor-pointer"
        @click="toggleColorMode"
      >
        <UIcon
          :name="colorMode.value === 'dark' ? 'i-lucide-sun' : 'i-lucide-moon'"
          class="size-4"
        />
      </button>
    </div>

    <!-- Ambient glowing backdrop -->
    <div class="pointer-events-none absolute -top-40 left-1/2 -translate-x-1/2 size-[600px] rounded-full bg-gradient-to-tr from-brand-600/10 dark:from-brand-600/15 via-primary-500/10 dark:via-primary-500/15 to-transparent blur-[120px]" />
    <div class="pointer-events-none absolute -bottom-40 right-1/4 size-[500px] rounded-full bg-gradient-to-br from-emerald-500/10 via-cyan-500/10 to-transparent blur-[120px]" />

    <div class="relative z-10 w-full max-w-md">
      <!-- Logo and Header -->
      <div class="mb-8 text-center">
        <div class="inline-flex size-14 items-center justify-center rounded-2xl border border-slate-200/80 dark:border-white/10 bg-white dark:bg-white/[0.04] shadow-lg dark:shadow-2xl backdrop-blur-md ring-1 ring-black/5 dark:ring-white/5 mb-4">
          <PolyMark class="size-8" />
        </div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-white sm:text-3xl font-display">
          Poly Cloud
        </h1>
        <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
          Sign in to access your unified multi-cloud storage
        </p>
      </div>

      <!-- Main Login Card -->
      <div class="rounded-2xl border border-slate-200/80 dark:border-white/[0.08] bg-white/95 dark:bg-[#0f131c]/85 p-7 shadow-xl dark:shadow-2xl backdrop-blur-xl ring-1 ring-black/5 dark:ring-black/40">
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <!-- Error banner -->
          <div
            v-if="errorMessage"
            class="flex items-center gap-2.5 rounded-xl border border-rose-500/25 bg-rose-500/10 px-3.5 py-2.5 text-xs text-rose-600 dark:text-rose-300"
          >
            <UIcon name="i-lucide-alert-circle" class="size-4 shrink-0 text-rose-500 dark:text-rose-400" />
            <span>{{ errorMessage }}</span>
          </div>

          <!-- Email -->
          <div class="space-y-1.5">
            <label class="block text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-300">
              Email Address
            </label>
            <div class="relative">
              <input
                v-model="email"
                type="email"
                autocomplete="email"
                required
                placeholder="admin@polycloud.local"
                class="w-full rounded-xl border border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/[0.03] px-3.5 py-2.5 pl-10 text-sm text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 transition-colors focus:border-primary-500 focus:bg-white dark:focus:bg-white/[0.05] focus:outline-none focus:ring-1 focus:ring-primary-500"
              />
              <UIcon
                name="i-lucide-mail"
                class="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-slate-400"
              />
            </div>
          </div>

          <!-- Password -->
          <div class="space-y-1.5">
            <label class="block text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-300">
              Password
            </label>
            <div class="relative">
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                required
                placeholder="••••••••"
                class="w-full rounded-xl border border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/[0.03] px-3.5 py-2.5 pl-10 pr-10 text-sm text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 transition-colors focus:border-primary-500 focus:bg-white dark:focus:bg-white/[0.05] focus:outline-none focus:ring-1 focus:ring-primary-500"
              />
              <UIcon
                name="i-lucide-lock"
                class="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-slate-400"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
                tabindex="-1"
              >
                <UIcon
                  :name="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'"
                  class="size-4"
                />
              </button>
            </div>
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="isLoading"
            class="mt-2 flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-brand-500 to-primary-600 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-brand-500/20 transition-all hover:brightness-110 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-60 cursor-pointer"
          >
            <UIcon
              v-if="isLoading"
              name="i-lucide-loader-2"
              class="size-4 animate-spin text-white"
            />
            <span class="text-white font-semibold">{{ isLoading ? 'Signing In...' : 'Sign In to Dashboard' }}</span>
          </button>
        </form>

        <!-- Default Credentials Callout -->
        <div class="mt-6 rounded-xl border border-slate-200/80 dark:border-white/[0.06] bg-slate-50 dark:bg-white/[0.02] p-3.5">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2 text-xs font-medium text-slate-700 dark:text-slate-300">
              <UIcon name="i-lucide-key-round" class="size-3.5 text-brand-500 dark:text-brand-400" />
              <span>Default Credentials</span>
            </div>
            <button
              type="button"
              @click="fillDefaultCredentials"
              class="text-[11px] font-medium text-brand-600 dark:text-brand-400 hover:underline underline-offset-2 transition-colors cursor-pointer"
            >
              Fill Credentials
            </button>
          </div>
          <div class="mt-2 flex flex-col gap-1 text-[11px] text-slate-500 dark:text-slate-400 font-mono">
            <div class="flex items-center justify-between">
              <span class="text-slate-400 dark:text-slate-500">Email:</span>
              <span class="text-slate-800 dark:text-slate-300 select-all font-semibold">admin@polycloud.local</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-slate-400 dark:text-slate-500">Password:</span>
              <span class="text-slate-800 dark:text-slate-300 select-all font-semibold">admin123</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer Info -->
      <div class="mt-6 text-center text-xs text-slate-500 dark:text-slate-500">
        <p>Configurable via <code class="rounded bg-slate-200/70 dark:bg-white/5 px-1 py-0.5 text-slate-700 dark:text-slate-400">ADMIN_EMAIL</code> & <code class="rounded bg-slate-200/70 dark:bg-white/5 px-1 py-0.5 text-slate-700 dark:text-slate-400">ADMIN_PASSWORD</code> in your environment.</p>
      </div>
    </div>
  </div>
</template>

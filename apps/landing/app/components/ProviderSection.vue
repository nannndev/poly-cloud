<script setup lang="ts">
import { GITHUB } from '~/config/site'

interface ProviderItem {
  name: string
  auth: string
  badge?: string
  status: 'active' | 'upcoming'
  iconKey: 'gdrive' | 'onedrive' | 'dropbox' | 's3' | 'r2' | 'b2' | 'nextcloud' | 'minio' | 'proton'
}

const activeProviders: ProviderItem[] = [
  { name: 'Google Drive', auth: 'OAuth 2.0', status: 'active', iconKey: 'gdrive' },
  { name: 'OneDrive', auth: 'OAuth 2.0', status: 'active', iconKey: 'onedrive' },
  { name: 'Dropbox', auth: 'OAuth 2.0', status: 'active', iconKey: 'dropbox' },
  { name: 'Amazon S3', auth: 'Access key', status: 'active', iconKey: 's3' },
  { name: 'Cloudflare R2', auth: 'Access key', status: 'active', iconKey: 'r2' },
  { name: 'Backblaze B2', auth: 'Access key', status: 'active', iconKey: 'b2' }
]

const upcomingProviders: ProviderItem[] = [
  { name: 'Nextcloud & WebDAV', auth: 'Self-Hosted', badge: 'Soon', status: 'upcoming', iconKey: 'nextcloud' },
  { name: 'MinIO & Wasabi', auth: 'S3-Compatible', badge: 'Soon', status: 'upcoming', iconKey: 'minio' },
  { name: 'Proton Drive', auth: 'E2EE Cloud', badge: 'Planned', status: 'upcoming', iconKey: 'proton' }
]
</script>

<template>
  <section id="providers" class="border-b border-white/[0.06] bg-ink-900 scroll-mt-16">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <div class="grid gap-10 lg:grid-cols-[0.85fr_1.15fr] lg:items-start lg:gap-16">
        <!-- Text Column -->
        <div>
          <div class="inline-flex items-center gap-2 rounded-full border border-white/[0.08] bg-ink-850 px-3 py-1 text-[11px] font-medium text-zinc-400">
            <span class="size-1.5 rounded-full bg-accent-400" />
            Universal Cloud Aggregator
          </div>

          <h2 class="mt-4 text-2xl font-semibold tracking-tight text-white md:text-3xl">
            6 providers today. Endless possibilities tomorrow.
          </h2>
          <p class="mt-4 max-w-[52ch] text-[15px] leading-relaxed text-zinc-400">
            OAuth providers connect smoothly through browser consent screens, while key-based
            providers need only an endpoint and API key. Once linked, all storage quotas and files
            are unified into one seamless namespace.
          </p>
          <p class="mt-4 max-w-[52ch] text-sm leading-relaxed text-zinc-500">
            Powered by the robust <a href="https://rclone.org" target="_blank" rel="noopener" class="text-zinc-400 hover:text-white underline">rclone</a> engine, Poly Cloud is architecturally decoupled. Adding new cloud providers requires zero database schema rewrites.
          </p>

          <div class="mt-8 rounded-2xl border border-white/[0.08] bg-ink-850 p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-xs font-semibold text-white">Need another provider?</p>
                <p class="text-[11px] text-zinc-400 mt-0.5">Vote on feature requests or contribute a driver.</p>
              </div>
              <a
                :href="GITHUB.issues"
                target="_blank"
                rel="noopener"
                class="inline-flex items-center gap-1.5 rounded-xl bg-ink-800 border border-white/[0.08] px-3 py-1.5 text-xs font-medium text-accent-300 hover:text-white hover:border-accent-400/40 transition-colors"
              >
                Suggest Provider ↗
              </a>
            </div>
          </div>
        </div>

        <!-- Providers Grid -->
        <div class="space-y-4">
          <div class="flex items-center justify-between px-1">
            <span class="text-[11px] font-bold uppercase tracking-wider text-zinc-500">Available Storage Providers</span>
            <span class="text-[11px] font-mono text-zinc-400">6 Connected • More in Roadmap</span>
          </div>

          <div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
            <!-- Active 6 Providers with Crisp Inline SVGs -->
            <div
              v-for="p in activeProviders"
              :key="p.name"
              class="group rounded-2xl border border-white/[0.08] bg-ink-850 px-4 py-5 text-center transition-all hover:border-accent-500/30 hover:bg-ink-800"
            >
              <div class="mx-auto flex size-8 items-center justify-center text-zinc-400 group-hover:text-white transition-colors">
                <!-- Google Drive -->
                <svg v-if="p.iconKey === 'gdrive'" viewBox="0 0 24 24" class="size-6" fill="currentColor">
                  <path d="M12.01 1.485c-2.082 0-3.754.02-3.743.047c.01.02 1.708 3.001 3.774 6.62l3.76 6.574h3.76c2.081 0 3.753-.02 3.742-.047c-.005-.02-1.708-3.001-3.775-6.62l-3.76-6.574zm-4.76 1.73a789.828 789.861 0 0 0-3.63 6.319L0 15.868l1.89 3.298l1.885 3.297l3.62-6.335l3.618-6.33l-1.88-3.287C8.1 4.704 7.255 3.22 7.25 3.214zm2.259 12.653l-.203.348c-.114.198-.96 1.672-1.88 3.287a423.93 423.948 0 0 1-1.698 2.97c-.01.026 3.24.042 7.222.042h7.244l1.796-3.157c.992-1.734 1.85-3.23 1.906-3.323l.104-.167h-7.249z"/>
                </svg>

                <!-- Microsoft OneDrive -->
                <svg v-else-if="p.iconKey === 'onedrive'" viewBox="0 0 24 24" class="size-6" fill="currentColor">
                  <path d="M19.453 9.95q.961.058 1.787.468t1.442 1.066q.615.657.966 1.512q.352.856.352 1.816q0 1.008-.387 1.893q-.386.885-1.049 1.547t-1.546 1.049q-.885.387-1.893.387H6q-1.242 0-2.332-.475t-1.904-1.29t-1.29-1.903Q0 14.93 0 13.688q0-.985.31-1.887q.311-.903.862-1.658q.55-.756 1.324-1.325q.774-.568 1.711-.861q.434-.129.85-.187q.416-.06.861-.082h.012q.515-.786 1.207-1.413q.691-.627 1.5-1.066q.808-.44 1.705-.668q.896-.229 1.845-.229q1.278 0 2.456.417q1.177.416 2.144 1.16t1.658 1.78q.692 1.038 1.008 2.28zm-7.265-4.137q-1.325 0-2.52.544q-1.195.545-2.04 1.565q.446.117.85.299q.405.181.792.416l4.78 2.86l2.731-1.15q.27-.117.545-.204q.276-.088.58-.147q-.293-.937-.855-1.705t-1.319-1.318q-.755-.551-1.658-.856t-1.886-.304M2.414 16.395l9.914-4.184l-3.832-2.297q-.586-.351-1.23-.539q-.645-.188-1.325-.188q-.914 0-1.722.364q-.809.363-1.412.978q-.604.616-.955 1.436q-.352.82-.352 1.723q0 .703.234 1.423t.68 1.284m16.711 1.793q.563 0 1.078-.176t.961-.516l-7.23-4.324l-10.301 4.336q.527.328 1.13.504q.604.175 1.237.175zm3.012-1.852q.363-.727.363-1.523q0-.774-.293-1.407t-.791-1.072t-1.166-.68t-1.406-.24q-.422 0-.838.1t-.815.252q-.398.152-.785.334q-.386.181-.761.345Z"/>
                </svg>

                <!-- Dropbox -->
                <svg v-else-if="p.iconKey === 'dropbox'" viewBox="0 0 24 24" class="size-6" fill="currentColor">
                  <path d="M6 1.807L0 5.629l6 3.822l6.001-3.822zm12 0l-6 3.822l6 3.822l6-3.822zM0 13.274l6 3.822l6.001-3.822L6 9.452zm18-3.822l-6 3.822l6 3.822l6-3.822zM6 18.371l6.001 3.822l6-3.822l-6-3.822z"/>
                </svg>

                <!-- Amazon S3 -->
                <svg v-else-if="p.iconKey === 's3'" viewBox="0 0 24 24" class="size-6" fill="currentColor">
                  <path d="m20.913 13.147l.12-.895c.947.576 1.258.922 1.354 1.071c-.16.031-.562.046-1.474-.176m-2.174 7.988l-.005.073c0 .084-.207.405-1.124.768a10 10 0 0 1-1.438.432c-1.405.325-3.128.504-4.853.504c-4.612 0-7.412-1.184-7.412-1.704l-.005-.073L1.81 5.602q.203.117.432.227q.064.03.128.057q.2.093.417.18l.179.069q.232.087.478.168l.13.043q.311.097.646.187l.176.044q.263.066.534.127a23 23 0 0 0 .843.17l.121.023q.378.067.768.122q.107.016.216.03q.299.04.604.077l.24.027q.366.04.74.07l.081.009q.413.033.83.056l.233.012q.316.015.633.025a33 33 0 0 0 2.795-.026l.232-.011q.417-.023.83-.056l.08-.008q.375-.03.742-.072l.238-.026q.307-.036.609-.077l.211-.03q.392-.056.772-.122l.111-.02q.323-.06.634-.125l.212-.047q.279-.062.546-.13l.166-.042q.338-.09.654-.189l.115-.038a11 11 0 0 0 .493-.173q.087-.032.17-.066q.225-.089.43-.185q.059-.025.116-.052q.23-.11.436-.228l-.976 7.245c-2.488-.78-5.805-2.292-7.311-3a1.09 1.09 0 0 0-1.088-1.085c-.6 0-1.088.489-1.088 1.088s.488 1.089 1.088 1.089c.196 0 .378-.056.537-.148c1.72.812 5.144 2.367 7.715 3.15zm-7.42-20.047c5.677 0 9.676 1.759 9.75 2.736l-.014.113c-.01.033-.031.067-.048.101c-.015.028-.026.057-.047.087c-.024.033-.058.068-.09.102c-.028.03-.051.06-.084.09c-.038.035-.087.07-.133.105c-.04.03-.074.06-.119.091c-.053.036-.116.071-.177.107c-.05.03-.095.06-.15.09c-.068.036-.147.073-.222.11c-.059.028-.114.057-.177.085c-.084.038-.177.074-.268.111c-.068.027-.13.054-.203.082c-.097.036-.205.072-.31.107c-.075.026-.148.053-.228.079c-.111.035-.233.069-.35.103c-.085.024-.165.05-.253.073c-.124.034-.258.065-.389.098c-.093.022-.181.046-.278.068c-.139.032-.287.061-.433.091c-.098.02-.191.041-.293.06c-.155.03-.32.057-.482.084c-.1.018-.198.036-.302.052c-.166.026-.342.048-.515.072c-.11.014-.213.03-.325.044c-.181.023-.372.041-.56.06q-.163.019-.332.036c-.188.016-.386.029-.58.043c-.122.009-.24.02-.364.028c-.207.012-.422.02-.635.028c-.12.005-.234.012-.354.016a36 36 0 0 1-2.069 0c-.12-.004-.234-.011-.352-.016c-.214-.008-.43-.016-.637-.028c-.122-.008-.238-.02-.36-.027c-.195-.015-.394-.028-.584-.044c-.11-.01-.215-.024-.324-.035c-.19-.02-.384-.038-.568-.06l-.315-.044c-.176-.024-.355-.046-.525-.073c-.1-.015-.192-.033-.29-.05c-.167-.028-.335-.055-.494-.086c-.096-.018-.183-.038-.276-.056c-.151-.032-.305-.062-.45-.095c-.09-.02-.173-.043-.26-.064c-.138-.034-.277-.067-.407-.102c-.082-.022-.157-.046-.235-.069a12 12 0 0 1-.368-.108c-.075-.024-.141-.049-.213-.073c-.11-.037-.223-.075-.325-.113c-.067-.025-.125-.051-.188-.077c-.096-.038-.195-.076-.282-.115c-.06-.027-.11-.054-.166-.08c-.08-.039-.162-.077-.233-.116c-.052-.028-.094-.055-.142-.084c-.063-.038-.13-.075-.185-.113c-.043-.029-.075-.058-.113-.086c-.048-.037-.098-.073-.139-.11c-.032-.029-.054-.057-.08-.087c-.033-.035-.069-.07-.093-.104c-.02-.03-.031-.058-.046-.086c-.018-.035-.039-.068-.049-.102l-.015-.113c.076-.977 4.074-2.736 9.748-2.736m12.182 12.124c-.118-.628-.84-1.291-2.31-2.128l.963-7.16l.005-.073C22.16 1.581 16.447 0 11.32 0C6.194 0 .482 1.581.482 3.851l.005.072L2.819 21.25c.071 2.002 5.236 2.75 8.5 2.75c1.805 0 3.615-.188 5.098-.531c.598-.138 1.133-.3 1.592-.48c1.18-.467 1.789-1.053 1.813-1.739l.945-7.018c.557.131 1.016.197 1.389.197c.54 0 .902-.137 1.134-.413a.96.96 0 0 0 .21-.804Z"/>
                </svg>

                <!-- Cloudflare R2 -->
                <svg v-else-if="p.iconKey === 'r2'" viewBox="0 0 24 24" class="size-6" fill="currentColor">
                  <path d="M16.509 16.845c.147-.507.09-.971-.155-1.316c-.225-.316-.605-.499-1.062-.52l-8.66-.113a.16.16 0 0 1-.133-.07a.2.2 0 0 1-.02-.156a.24.24 0 0 1 .203-.156l8.736-.113c1.035-.049 2.16-.886 2.554-1.913l.499-1.302a.27.27 0 0 0 .014-.168a5.689 5.689 0 0 0-10.937-.584a2.58 2.58 0 0 0-1.794-.498a2.56 2.56 0 0 0-2.223 3.18A3.634 3.634 0 0 0 0 16.751q.002.264.035.527a.174.174 0 0 0 .17.148h15.98a.22.22 0 0 0 .204-.155zm2.757-5.564c-.077 0-.161 0-.239.011c-.056 0-.105.042-.127.098l-.337 1.174c-.148.507-.092.971.154 1.317c.225.316.605.498 1.062.52l1.844.113c.056 0 .105.026.133.07a.2.2 0 0 1 .021.156a.24.24 0 0 1-.204.156l-1.92.112c-1.042.049-2.159.887-2.553 1.914l-.141.358c-.028.072.021.142.099.142h6.597a.174.174 0 0 0 .17-.126a5 5 0 0 0 .175-1.28a4.74 4.74 0 0 0-4.734-4.727"/>
                </svg>

                <!-- Backblaze B2 -->
                <svg v-else-if="p.iconKey === 'b2'" viewBox="0 0 24 24" class="size-6" fill="currentColor">
                  <path d="M9.31 0c.653 1.35 1.567 4.082-1.388 7.174c-1.81 1.88-3.078 3.849-2.35 6.065c.365 1.103 1.187 2.507 2.887 2.785c.61.1 1.343 0 1.74-.14c2.454-.855 2.098-3.415 1.555-5.048c-.07-.213-.191-.733-.236-.924c-.373-1.602.776-2.656 1.129-3.804q.043-.138.07-.272q.062-.315.078-.638c0-1.827-.988-2.63-1.775-3.6C10.18.564 9.31 0 9.31 0m6.276 6.018s-.709.336-1.219.883c-.445.482-.863.879-1.294 1.859q-.041.21-.075.438c-.232 1.641 1.148 3.144.719 5.189c-.112.535-.355.712-.781 1.637c-.51 1.106-.383 2.588.36 3.529c.672.849 1.878 1.232 3.052.95c2.106-.505 3.065-2.283 2.896-4.286c-.131-1.58-.815-2.753-2.754-4.96c-.96-1.093-1.607-2.41-1.562-3.407c.137-1.207.658-1.832.658-1.832M4.893 15.194c-.022.014-.044.061-.059.16l-.006.02v.01c-.114.54-.165 1.822.116 2.968c.353 1.443 1.417 3.902 4.412 5.129c2.518 1.034 5.718.541 7.85-1.627c.529-.543.407-.49-.489-.201v-.002c-1.112.356-3.518.546-4.768-1c-1.523-1.885-.43-3.363-1.357-3.15c-3.616.834-5.267-1.466-5.547-2.102c-.002-.002-.086-.249-.152-.205"/>
                </svg>
              </div>
              <p class="mt-3 text-[13px] font-semibold text-zinc-200">{{ p.name }}</p>
              <p class="mt-0.5 text-[11px] text-zinc-500">{{ p.auth }}</p>
            </div>
          </div>

          <!-- Section: Coming Soon & Expandable Providers -->
          <div class="pt-2">
            <div class="flex items-center justify-between px-1 mb-3">
              <span class="text-[11px] font-bold uppercase tracking-wider text-accent-400">Roadmap &amp; Upcoming</span>
              <span class="text-[11px] text-zinc-500">Easily integrated via rclone</span>
            </div>

            <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <div
                v-for="p in upcomingProviders"
                :key="p.name"
                class="rounded-2xl border border-white/[0.06] bg-ink-950/60 p-3.5 text-center relative overflow-hidden"
              >
                <span class="absolute top-2 right-2 rounded-md bg-accent-500/15 border border-accent-400/25 px-1.5 py-0.2 text-[9px] font-bold uppercase tracking-wider text-accent-300">
                  {{ p.badge }}
                </span>

                <div class="mx-auto flex size-7 items-center justify-center text-zinc-500">
                  <!-- Nextcloud / WebDAV -->
                  <svg v-if="p.iconKey === 'nextcloud'" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
                    <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
                    <line x1="6" y1="6" x2="6.01" y2="6" />
                    <line x1="6" y1="18" x2="6.01" y2="18" />
                  </svg>

                  <!-- MinIO / Wasabi -->
                  <svg v-else-if="p.iconKey === 'minio'" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242" />
                    <path d="M12 12v9" />
                    <path d="m8 17 4 4 4-4" />
                  </svg>

                  <!-- Proton Drive -->
                  <svg v-else viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                  </svg>
                </div>

                <p class="mt-2 text-xs font-semibold text-zinc-300">{{ p.name }}</p>
                <p class="mt-0.5 text-[10px] text-zinc-500">{{ p.auth }}</p>
              </div>

              <!-- Interactive Request Provider Button -->
              <a
                :href="GITHUB.issues"
                target="_blank"
                rel="noopener"
                class="group flex flex-col items-center justify-center rounded-2xl border border-dashed border-white/[0.14] bg-ink-850/40 p-3.5 text-center transition-all hover:border-accent-400/50 hover:bg-ink-800"
              >
                <div class="flex size-7 items-center justify-center rounded-full bg-accent-500/10 text-accent-400 group-hover:scale-110 transition-transform">
                  <svg viewBox="0 0 24 24" class="size-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="12" y1="5" x2="12" y2="19" />
                    <line x1="5" y1="12" x2="19" y2="12" />
                  </svg>
                </div>
                <p class="mt-2 text-xs font-bold text-accent-300 group-hover:text-white transition-colors">Suggest More</p>
                <p class="mt-0.5 text-[10px] text-zinc-500">via GitHub ↗</p>
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

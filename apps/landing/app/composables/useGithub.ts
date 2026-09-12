import { REPO_NAME, REPO_OWNER } from '~/config/site'

export interface Contributor {
  login: string
  avatar: string
  url: string
  contributions: number
  /** Akun bot (dependabot dkk.) disaring: mereka bukan orang. */
  isBot: boolean
}

export interface RepoStats {
  stars: number
  forks: number
  openIssues: number
  license: string | null
  pushedAt: string | null
}

const API = 'https://api.github.com'

/**
 * Permintaan tak terautentikasi ke GitHub dibatasi 60/jam per IP. Halaman ini
 * dirender saat build, jadi kuota itu cukup — tapi build juga harus tetap jalan
 * ketika batas terlampaui atau GitHub sedang bermasalah. Setiap pengambilan di
 * bawah karena itu punya nilai bawaan, dan kegagalannya tak pernah
 * menggagalkan build.
 */
const headers = { Accept: 'application/vnd.github+json' }

export function useRepoStats() {
  return useAsyncData<RepoStats>('gh-repo', async () => {
    const repo = await $fetch<any>(`${API}/repos/${REPO_OWNER}/${REPO_NAME}`, { headers })
    return {
      stars: repo.stargazers_count ?? 0,
      forks: repo.forks_count ?? 0,
      openIssues: repo.open_issues_count ?? 0,
      license: repo.license?.spdx_id ?? null,
      pushedAt: repo.pushed_at ?? null
    }
  }, {
    default: (): RepoStats => ({
      stars: 0, forks: 0, openIssues: 0, license: null, pushedAt: null
    })
  })
}

export function useContributors() {
  return useAsyncData<Contributor[]>('gh-contributors', async () => {
    const raw = await $fetch<any[]>(
      `${API}/repos/${REPO_OWNER}/${REPO_NAME}/contributors`,
      { headers, query: { per_page: 100, anon: 0 } }
    )
    return (raw ?? [])
      .map(c => ({
        login: c.login as string,
        // Avatar diminta pada ukuran render agar tak mengunduh 460px untuk
        // lingkaran 48px.
        avatar: `${c.avatar_url}&s=160`,
        url: c.html_url as string,
        contributions: c.contributions ?? 0,
        isBot: c.type === 'Bot' || /\[bot\]$/.test(c.login ?? '')
      }))
      .filter(c => !c.isBot)
      .sort((a, b) => b.contributions - a.contributions)
  }, { default: (): Contributor[] => [] })
}

/** Angka besar diringkas (1200 → 1.2k) supaya tak memaksa kartu melebar. */
export function formatCount(n: number): string {
  if (n < 1000) return String(n)
  return `${(n / 1000).toFixed(n < 10_000 ? 1 : 0).replace(/\.0$/, '')}k`
}

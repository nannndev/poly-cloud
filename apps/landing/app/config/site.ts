/**
 * Satu tempat untuk setiap tautan luar yang dipakai landing page. Ditaruh di
 * sini, bukan disebar di komponen, supaya mengganti username atau menambah
 * kanal donasi tak berarti menyisir seluruh berkas.
 */
export const REPO_OWNER = 'nannndev'
export const REPO_NAME = 'poly-cloud'
export const REPO_URL = `https://github.com/${REPO_OWNER}/${REPO_NAME}`

export const GITHUB = {
  repo: REPO_URL,
  issues: `${REPO_URL}/issues`,
  discussions: `${REPO_URL}/discussions`,
  contributing: `${REPO_URL}/blob/main/CONTRIBUTING.md`,
  goodFirstIssues: `${REPO_URL}/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22`,
  releases: `${REPO_URL}/releases`,
  license: `${REPO_URL}/blob/main/LICENSE`
}

export interface DonationChannel {
  id: string
  name: string
  /** Kalimat yang menjelaskan kapan kanal ini paling masuk akal dipilih. */
  note: string
  url: string
  /** slug simpleicons.org; kosong berarti pakai ikon gambar sendiri. */
  icon: string
  /** Warna merek untuk ikon, dipakai agar tiap kanal mudah dibedakan. */
  tint: string
  region: 'global' | 'id'
  /**
   * false = kanal disembunyikan dari halaman.
   *
   * Tautan donasi yang mati lebih buruk daripada tak ada sama sekali: orang yang
   * berniat menyumbang berakhir di halaman 404. Kanal yang akunnya belum ada
   * dimatikan di sini, bukan dihapus, supaya tinggal dinyalakan begitu siap.
   */
  enabled: boolean
}

/**
 * Urutan sengaja: GitHub Sponsors lebih dulu karena paling dekat dengan
 * kebiasaan penyumbang proyek sumber terbuka, lalu kanal sekali-bayar, lalu
 * kanal lokal yang menerima QRIS/e-wallet tanpa kartu kredit.
 *
 * Tautan di bawah memakai username pemilik repo. Ganti bila akun pembayaran
 * memakai nama lain — halaman akan mengikuti tanpa perubahan lain.
 *
 * Status per 13 September 2026: GitHub Sponsors dan Saweria terverifikasi aktif.
 * PayPal ("isn't available") dan Trakteer (404) belum ada untuk handle ini, jadi
 * enabled:false. Nyalakan setelah akunnya dibuat, atau ganti `url`-nya bila
 * handle-nya berbeda dari username GitHub.
 */
export const DONATIONS: DonationChannel[] = [
  {
    id: 'github',
    name: 'GitHub Sponsors',
    note: 'Recurring or one-off, billed with your existing GitHub account.',
    url: `https://github.com/sponsors/${REPO_OWNER}`,
    icon: 'github',
    tint: '#f4f4f5',
    region: 'global',
    enabled: true
  },
  {
    id: 'kofi',
    name: 'Ko-fi',
    note: 'One-off tip with a card or PayPal. No account needed.',
    url: `https://ko-fi.com/${REPO_OWNER}`,
    icon: 'kofi',
    tint: '#ff5e5b',
    region: 'global',
    enabled: true
  },
  {
    id: 'paypal',
    name: 'PayPal',
    note: 'Direct transfer, any amount, any currency.',
    url: `https://paypal.me/${REPO_OWNER}`,
    icon: 'paypal',
    tint: '#0079c1',
    region: 'global',
    enabled: false
  },
  {
    id: 'saweria',
    name: 'Saweria',
    note: 'QRIS and Indonesian e-wallets. No card required.',
    url: `https://saweria.co/${REPO_OWNER}`,
    icon: '',
    tint: '#fbbf24',
    region: 'id',
    enabled: true
  },
  {
    id: 'trakteer',
    name: 'Trakteer',
    note: 'Indonesian bank transfer, e-wallet, or QRIS.',
    url: `https://trakteer.id/${REPO_OWNER}`,
    icon: '',
    tint: '#f87171',
    region: 'id',
    enabled: false
  }
]

/**
 * Alamat kripto hanya ditampilkan bila diisi. Dibiarkan kosong secara bawaan:
 * memajang alamat contoh pada halaman donasi berisiko menyesatkan — uang bisa
 * terkirim ke dompet yang bukan milik siapa pun di proyek ini.
 */
export const CRYPTO: { symbol: string; network: string; address: string }[] = []

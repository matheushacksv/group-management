// Path to the Go API through the Nuxt proxy, scoped to the OpenWA session
export function groupsPath(groupId?: string) {
  const { sessionId } = useRuntimeConfig().public
  const base = `/api/go/groups/${sessionId}`
  return groupId ? `${base}/${groupId}` : base
}

export function formatDate(iso?: string) {
  return iso ? new Date(iso).toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' }) : '-'
}

const rtf = new Intl.RelativeTimeFormat('pt-BR', { numeric: 'auto' })
const units: [Intl.RelativeTimeFormatUnit, number][] = [
  ['year', 31_536_000],
  ['month', 2_592_000],
  ['day', 86_400],
  ['hour', 3_600],
  ['minute', 60]
]

// "há 3 meses", "ontem", "agora"; minute is the smallest unit so SSR and browser render the same text
export function timeAgo(iso: string) {
  const seconds = (new Date(iso).getTime() - Date.now()) / 1000
  for (const [unit, size] of units) {
    if (Math.abs(seconds) >= size) {
      return rtf.format(Math.round(seconds / size), unit)
    }
  }
  return 'agora'
}

export function apiError(err: unknown) {
  const data = (err as { data?: { error?: string } })?.data
  return data?.error ?? 'Erro inesperado'
}

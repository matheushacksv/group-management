// Removals still running, shared by key: sidebar badges and the "finished" notice read the same data
export function useActiveRemovals() {
  return useFetch<ActiveRemoval[]>('/api/go/removals/active', {
    key: 'active-removals',
    default: () => []
  })
}

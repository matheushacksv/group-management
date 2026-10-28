// Stored groups shared by key: the sidebar and the groups page read the same data,
// so marking a group as managed updates both at once
export function useGroups() {
  // deep: the managed switch mutates the group in place and the sidebar must react to it
  return useFetch<StoredGroup[]>(groupsPath(), { key: 'groups', default: () => [], deep: true })
}

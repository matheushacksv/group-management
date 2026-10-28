<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

const toast = useToast()

const { data: groups, status, refresh } = await useGroups()

// useState keeps the filters when navigating to a group and back
const search = useState('groups-search', () => '')
const onlyManaged = useState('groups-only-managed', () => false)

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  return groups.value.filter(g =>
    (!onlyManaged.value || g.managed) && g.name.toLowerCase().includes(q)
  )
})

const managedCount = computed(() => groups.value.filter(g => g.managed).length)

const columns: TableColumn<StoredGroup>[] = [
  { accessorKey: 'name', header: 'Grupo' },
  { accessorKey: 'managed', header: 'Gerenciado' },
  { accessorKey: 'syncedAt', header: 'Sincronizado' },
  { id: 'actions' }
]

const syncing = ref(false)
async function sync() {
  syncing.value = true
  try {
    const res = await $fetch<{ saved: number }>(`${groupsPath()}/sync`, { method: 'POST' })
    toast.add({ title: `${res.saved} grupos sincronizados`, color: 'success', icon: 'i-lucide-check' })
    await refresh()
  } catch (err) {
    toast.add({ title: 'Falha ao sincronizar', description: apiError(err), color: 'error' })
  } finally {
    syncing.value = false
  }
}

async function setManaged(group: StoredGroup, managed: boolean) {
  const previous = group.managed
  group.managed = managed
  try {
    await $fetch(groupsPath(group.id), { method: 'PATCH', body: { managed } })
  } catch (err) {
    group.managed = previous
    toast.add({ title: 'Falha ao atualizar', description: apiError(err), color: 'error' })
    return
  }

  if (managed) {
    await syncGroup(group)
  }
}

// loads the members of a single group right after it becomes managed
const syncingGroup = ref<string>()
async function syncGroup(group: StoredGroup) {
  syncingGroup.value = group.id
  try {
    await $fetch(`${groupsPath(group.id)}/sync`, { method: 'POST' })
    toast.add({ title: 'Membros carregados', description: group.name, color: 'success', icon: 'i-lucide-check' })
    await refresh()
  } catch (err) {
    toast.add({ title: 'Grupo marcado, mas falhou ao carregar membros', description: apiError(err), color: 'warning' })
  } finally {
    syncingGroup.value = undefined
  }
}
</script>

<template>
  <UDashboardPanel id="groups">
    <template #header>
      <UDashboardNavbar title="Grupos">
        <template #trailing>
          <UBadge :label="`${managedCount} de ${groups.length} gerenciados`" color="neutral" variant="subtle" />
        </template>
        <template #right>
          <UButton icon="i-lucide-refresh-cw" label="Sincronizar" :loading="syncing" @click="sync" />
        </template>
      </UDashboardNavbar>

      <UDashboardToolbar>
        <template #left>
          <UInput v-model="search" icon="i-lucide-search" placeholder="Buscar grupo" class="w-64" />
          <USwitch v-model="onlyManaged" label="Só gerenciados" />
        </template>
      </UDashboardToolbar>
    </template>

    <template #body>
      <UAlert
        v-if="syncing"
        icon="i-lucide-info"
        color="info"
        variant="subtle"
        title="Sincronizando com o WhatsApp"
        description="Grupos gerenciados levam ~1s cada. Não feche a página."
      />

      <UTable :data="filtered" :columns="columns" :loading="status === 'pending'" sticky>
        <template #name-cell="{ row }">
          <div class="flex items-center gap-3 min-w-0">
            <UAvatar icon="i-lucide-users" size="sm" :class="row.original.managed ? 'text-primary' : ''" />
            <span class="truncate font-medium" :class="row.original.closedAt ? 'text-muted line-through' : 'text-highlighted'">{{ row.original.name }}</span>
          </div>
        </template>
        <template #managed-cell="{ row }">
          <UTooltip v-if="row.original.closedAt" :text="`Encerrado ${timeAgo(row.original.closedAt)}`">
            <UBadge color="neutral" variant="subtle" icon="i-lucide-power" label="Encerrado" />
          </UTooltip>
          <USwitch
            v-else
            :model-value="row.original.managed"
            :loading="syncingGroup === row.original.id"
            :disabled="syncingGroup === row.original.id"
            @update:model-value="(v: boolean) => setManaged(row.original, v)"
          />
        </template>
        <template #syncedAt-cell="{ row }">
          <TimeAgo :date="row.original.syncedAt" class="text-muted" />
        </template>
        <template #actions-cell="{ row }">
          <div class="flex justify-end">
            <UButton
              v-if="row.original.managed || row.original.closedAt"
              :to="`/groups/${row.original.id}`"
              label="Abrir"
              trailing-icon="i-lucide-chevron-right"
              variant="ghost"
              size="sm"
            />
          </div>
        </template>
        <template #empty>
          <UEmpty icon="i-lucide-users" title="Nenhum grupo" description="Clique em Sincronizar para buscar os grupos." />
        </template>
      </UTable>
    </template>
  </UDashboardPanel>
</template>

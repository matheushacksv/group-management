<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

const groupId = useRoute().params.id as string

// group name comes from the stored list (shared with the sidebar), avoids a live call to WhatsApp
const { data: groups } = await useGroups()
const group = computed(() => groups.value.find(g => g.id === groupId))

const { data: returned, status: returnedStatus } = await useFetch<ReturnedMember[]>(
  `${groupsPath(groupId)}/members/returned`,
  { default: () => [], lazy: true }
)
const { data: jobs, status: jobsStatus } = await useFetch<RemovalJob[]>(
  `${groupsPath(groupId)}/removals`,
  { default: () => [], lazy: true }
)

const tabs = computed(() => [
  { label: 'Membros', icon: 'i-lucide-users', slot: 'members' as const },
  { label: 'Saíram e voltaram', icon: 'i-lucide-repeat', slot: 'returned' as const, badge: returned.value.length || undefined },
  { label: 'Remoções', icon: 'i-lucide-history', slot: 'removals' as const, badge: jobs.value.length || undefined }
])

const returnedColumns: TableColumn<ReturnedMember>[] = [
  { accessorKey: 'pushName', header: 'Membro' },
  { accessorKey: 'exits', header: 'Saídas' },
  { accessorKey: 'lastExit', header: 'Última saída' },
  { accessorKey: 'lastJoin', header: 'Voltou' }
]

const closeOpen = ref(false)
const actions = computed(() => [[
  { label: 'Encerrar grupo', icon: 'i-lucide-power', color: 'error' as const, onSelect: () => { closeOpen.value = true } }
]])

const jobColumns: TableColumn<RemovalJob>[] = [
  { accessorKey: 'id', header: '#' },
  { accessorKey: 'createdAt', header: 'Criada' },
  { accessorKey: 'requestedBy', header: 'Por' },
  { id: 'progress', header: 'Resultado' },
  { id: 'actions' }
]
</script>

<template>
  <UDashboardPanel :id="`group-${groupId}`">
    <template #header>
      <UDashboardNavbar :title="group?.name ?? groupId">
        <template #leading>
          <UButton to="/" icon="i-lucide-arrow-left" color="neutral" variant="ghost" aria-label="Voltar para grupos" />
        </template>
        <template #trailing>
          <UBadge v-if="group?.closedAt" color="neutral" variant="subtle" icon="i-lucide-power" label="Encerrado" />
          <UBadge v-else-if="group && !group.managed" color="warning" variant="subtle" label="Não gerenciado" />
        </template>
        <template #right>
          <UDropdownMenu v-if="group?.managed && !group.closedAt" :items="actions" :content="{ align: 'end' }">
            <UButton icon="i-lucide-ellipsis-vertical" color="neutral" variant="ghost" aria-label="Mais ações" />
          </UDropdownMenu>
        </template>
      </UDashboardNavbar>
      <CloseGroupModal v-if="group" v-model:open="closeOpen" :group="group" />
    </template>

    <template #body>
      <UTabs :items="tabs" variant="link" class="w-full">
        <template #members>
          <MembersTab :group-id="groupId" class="pt-4" />
        </template>

        <template #returned>
          <UTable :data="returned" :columns="returnedColumns" :loading="returnedStatus === 'pending'" class="pt-4">
            <template #pushName-cell="{ row }">
              <MemberCell :name="row.original.pushName" :number="row.original.number" />
            </template>
            <template #lastExit-cell="{ row }"><TimeAgo :date="row.original.lastExit" /></template>
            <template #lastJoin-cell="{ row }"><TimeAgo :date="row.original.lastJoin" /></template>
            <template #empty>
              <UEmpty icon="i-lucide-repeat" title="Ninguém saiu e voltou" />
            </template>
          </UTable>
        </template>

        <template #removals>
          <UTable :data="jobs" :columns="jobColumns" :loading="jobsStatus === 'pending'" class="pt-4">
            <template #createdAt-cell="{ row }"><TimeAgo :date="row.original.createdAt" /></template>
            <template #progress-cell="{ row }">
              <div class="flex gap-2">
                <UBadge v-if="row.original.pending" color="neutral" variant="subtle" icon="i-lucide-loader" :label="`${row.original.pending} pendente(s)`" />
                <UBadge v-if="row.original.ok" color="success" variant="subtle" icon="i-lucide-check" :label="`${row.original.ok} ok`" />
                <UBadge v-if="row.original.failed" color="error" variant="subtle" icon="i-lucide-x" :label="`${row.original.failed} falha(s)`" />
                <UBadge v-if="row.original.canceled" color="warning" variant="subtle" icon="i-lucide-ban" :label="`${row.original.canceled} cancelada(s)`" />
              </div>
            </template>
            <template #requestedBy-cell="{ row }">
              <span class="text-sm text-muted">{{ row.original.requestedBy ?? '-' }}</span>
            </template>
            <template #actions-cell="{ row }">
              <div class="flex justify-end">
                <UButton
                  :to="`/removals/${row.original.id}?group=${groupId}`"
                  label="Detalhes"
                  trailing-icon="i-lucide-chevron-right"
                  variant="ghost"
                  size="sm"
                />
              </div>
            </template>
            <template #empty>
              <UEmpty icon="i-lucide-history" title="Nenhuma remoção" />
            </template>
          </UTable>
        </template>
      </UTabs>
    </template>
  </UDashboardPanel>
</template>

<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

const props = defineProps<{ groupId: string }>()
const toast = useToast()

const minDays = ref(0)
const periods = [
  { label: 'Todos', value: 0 },
  { label: '+30 dias', value: 30 },
  { label: '+90 dias', value: 90 },
  { label: '+180 dias', value: 180 },
  { label: '+1 ano', value: 365 }
]

const { data: members, status } = await useFetch<Member[]>(
  () => `${groupsPath(props.groupId)}/members`,
  { query: { min_days: minDays }, default: () => [] }
)

const search = ref('')
const filtered = computed(() => {
  const q = search.value.toLowerCase()
  return members.value.filter(m =>
    m.number.includes(q) || (m.pushName ?? '').toLowerCase().includes(q)
  )
})

// selection survives filtering, cleared when the period changes
const selected = ref(new Set<string>())
watch(minDays, () => selected.value.clear())

const allSelected = computed(() =>
  filtered.value.length > 0 && filtered.value.every(m => selected.value.has(m.id))
)
function toggleAll(value: boolean) {
  for (const m of filtered.value) {
    if (value) selected.value.add(m.id)
    else selected.value.delete(m.id)
  }
}
function toggle(id: string, value: boolean) {
  if (value) selected.value.add(id)
  else selected.value.delete(id)
}

const columns: TableColumn<Member>[] = [
  { id: 'select' },
  { accessorKey: 'pushName', header: 'Membro' },
  { accessorKey: 'since', header: 'No grupo' }
]

// removal confirmation: user must type the amount, removal can't be undone
const confirmOpen = ref(false)
const confirmText = ref('')
const removing = ref(false)
const selectedMembers = computed(() => members.value.filter(m => selected.value.has(m.id)))

function openConfirm() {
  confirmText.value = ''
  confirmOpen.value = true
}

async function remove() {
  removing.value = true
  try {
    const res = await $fetch<CreateRemovalResponse>(`${groupsPath(props.groupId)}/removals`, {
      method: 'POST',
      body: { participants: [...selected.value] }
    })
    if (res.skipped.length) {
      toast.add({
        title: `${res.skipped.length} ignorado(s)`,
        description: res.skipped.map(s => `${s.id.replace('@c.us', '')}: ${s.reason}`).join(', '),
        color: 'warning'
      })
    }
    selected.value.clear()
    confirmOpen.value = false
    await navigateTo(`/removals/${res.jobId}?group=${props.groupId}`)
  } catch (err) {
    toast.add({ title: 'Falha ao criar remoção', description: apiError(err), color: 'error' })
  } finally {
    removing.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center gap-3">
      <UInput v-model="search" icon="i-lucide-search" placeholder="Buscar nome ou número" class="w-64" />
      <USelect v-model="minDays" :items="periods" class="w-40" />
      <span class="text-sm text-muted">{{ filtered.length }} membro(s)</span>
      <UButton
        class="ms-auto"
        icon="i-lucide-user-minus"
        color="error"
        :label="`Remover selecionados (${selected.size})`"
        :disabled="selected.size === 0"
        @click="openConfirm"
      />
    </div>

    <UAlert
      v-if="minDays > 0"
      icon="i-lucide-info"
      variant="subtle"
      color="neutral"
      description="Para quem já estava no grupo antes do sistema, a data é a do primeiro sync, não a entrada real."
    />

    <UTable :data="filtered" :columns="columns" :loading="status === 'pending'" sticky>
      <template #select-header>
        <UCheckbox :model-value="allSelected" @update:model-value="(v) => toggleAll(v === true)" />
      </template>
      <template #select-cell="{ row }">
        <UCheckbox
          :model-value="selected.has(row.original.id)"
          @update:model-value="(v) => toggle(row.original.id, v === true)"
        />
      </template>
      <template #pushName-cell="{ row }">
        <MemberCell :name="row.original.pushName" :number="row.original.number" />
      </template>
      <template #since-cell="{ row }">
        <TimeAgo :date="row.original.since" />
      </template>
      <template #empty>
        <UEmpty icon="i-lucide-users" title="Nenhum membro" description="Sincronize os grupos para carregar os membros." />
      </template>
    </UTable>

    <UModal v-model:open="confirmOpen" title="Confirmar remoção">
      <template #body>
        <div class="space-y-3">
          <p>
            Remover <strong>{{ selected.size }}</strong> pessoa(s) do grupo. A remoção é feita aos poucos e
            <strong>não pode ser desfeita</strong>: quem sai pode não conseguir ser adicionado de volta.
          </p>
          <ul class="max-h-40 overflow-y-auto text-sm text-muted">
            <li v-for="m in selectedMembers" :key="m.id">
              {{ m.pushName ?? m.number }} <span v-if="m.pushName">({{ m.number }})</span>
            </li>
          </ul>
          <UFormField :label="`Digite ${selected.size} para confirmar`">
            <UInput v-model="confirmText" class="w-full" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <UButton label="Cancelar" variant="ghost" color="neutral" @click="confirmOpen = false" />
        <UButton
          label="Remover"
          color="error"
          :loading="removing"
          :disabled="confirmText !== String(selected.size)"
          @click="remove"
        />
      </template>
    </UModal>
  </div>
</template>

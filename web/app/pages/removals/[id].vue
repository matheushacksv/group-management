<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

const route = useRoute()
const toast = useToast()
const jobId = route.params.id as string
const groupId = route.query.group as string | undefined

const { data: job, refresh, error } = await useFetch<RemovalDetail>(`/api/go/removals/${jobId}`)

const total = computed(() => job.value?.items.length ?? 0)
const pending = computed(() => job.value?.counts.pending ?? 0)
const canceled = computed(() => job.value?.counts.canceled ?? 0)
const done = computed(() => total.value - pending.value)

// poll while the worker is still processing
const timer = ref<ReturnType<typeof setInterval>>()
onMounted(() => {
  timer.value = setInterval(() => {
    if (pending.value > 0) refresh()
  }, 3000)
})
onUnmounted(() => clearInterval(timer.value))

// cancel stops the pending items; the one being removed right now still finishes
const cancelOpen = ref(false)
const canceling = ref(false)
async function cancel() {
  canceling.value = true
  try {
    const res = await $fetch<{ canceled: number }>(`/api/go/removals/${jobId}/cancel`, { method: 'POST' })
    toast.add({ title: `${res.canceled} remoção(ões) cancelada(s)`, color: 'success', icon: 'i-lucide-check' })
    cancelOpen.value = false
    await refresh()
  } catch (err) {
    toast.add({ title: 'Falha ao cancelar', description: apiError(err), color: 'error' })
  } finally {
    canceling.value = false
  }
}

const statusColor = { pending: 'neutral', ok: 'success', failed: 'error', canceled: 'warning' } as const
const statusLabel = { pending: 'Pendente', ok: 'Removido', failed: 'Falhou', canceled: 'Cancelado' } as const
const statusIcon = { pending: 'i-lucide-loader', ok: 'i-lucide-check', failed: 'i-lucide-x', canceled: 'i-lucide-ban' } as const

const stats = computed(() => [
  { label: 'Total', value: total.value, icon: 'i-lucide-users', color: 'text-default' },
  { label: 'Removidos', value: job.value?.counts.ok ?? 0, icon: 'i-lucide-check', color: 'text-success' },
  { label: 'Falhas', value: job.value?.counts.failed ?? 0, icon: 'i-lucide-x', color: 'text-error' },
  canceled.value
    ? { label: 'Cancelados', value: canceled.value, icon: 'i-lucide-ban', color: 'text-warning' }
    : { label: 'Pendentes', value: pending.value, icon: 'i-lucide-loader', color: 'text-muted' }
])

const columns: TableColumn<RemovalItem>[] = [
  { accessorKey: 'id', header: 'Número' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'error', header: 'Motivo' },
  { accessorKey: 'processedAt', header: 'Processado' }
]
</script>

<template>
  <UDashboardPanel :id="`removal-${jobId}`">
    <template #header>
      <UDashboardNavbar :title="job?.closeGroup ? `Encerramento #${jobId}` : `Remoção #${jobId}`">
        <template #leading>
          <UButton
            :to="groupId ? `/groups/${groupId}` : '/'"
            icon="i-lucide-arrow-left"
            color="neutral"
            variant="ghost"
            aria-label="Voltar"
          />
        </template>
        <template #trailing>
          <UBadge v-if="pending > 0" color="info" variant="subtle" icon="i-lucide-loader" label="Em andamento" />
          <UBadge v-else-if="canceled" color="warning" variant="subtle" icon="i-lucide-ban" label="Cancelada" />
          <UBadge v-else-if="job" color="success" variant="subtle" icon="i-lucide-check" label="Concluída" />
        </template>
        <template #right>
          <UButton
            v-if="pending > 0"
            icon="i-lucide-octagon-x"
            label="Cancelar pendentes"
            color="error"
            variant="soft"
            @click="cancelOpen = true"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <UAlert v-if="error" color="error" variant="subtle" title="Remoção não encontrada" />

      <template v-if="job">
        <p class="text-sm text-muted">
          Criada <TimeAgo :date="job.createdAt" />
          <template v-if="job.requestedBy"> por <span class="text-default">{{ job.requestedBy }}</span></template>
        </p>

        <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <UCard v-for="s in stats" :key="s.label" :ui="{ body: 'p-4 sm:p-4' }">
            <div class="flex items-center justify-between">
              <span class="text-sm text-muted">{{ s.label }}</span>
              <UIcon :name="s.icon" class="size-4" :class="s.color" />
            </div>
            <p class="mt-1 text-2xl font-semibold" :class="s.color">{{ s.value }}</p>
          </UCard>
        </div>

        <div class="space-y-2">
          <UProgress :model-value="done" :max="total" :color="job.counts.failed ? 'warning' : 'primary'" />
          <p v-if="pending > 0" class="text-sm text-muted">
            Processando 1 pessoa a cada 3 a 6 segundos para evitar bloqueio. Pode fechar a página, a fila continua.
          </p>
          <UAlert
            v-if="job.closeGroup"
            :color="canceled ? 'warning' : 'info'"
            variant="subtle"
            icon="i-lucide-power"
            :title="canceled ? 'Encerramento cancelado: o bot continua no grupo' : 'Encerramento do grupo'"
            :description="canceled
              ? 'Quem já saiu não volta. O grupo continua ativo com quem sobrou.'
              : 'Ao terminar a fila, o bot sai do grupo automaticamente e ele é marcado como encerrado.'"
          />
        </div>

        <UTable :data="job.items" :columns="columns">
          <template #id-cell="{ row }">
            <MemberCell :number="row.original.id.replace('@c.us', '')" />
          </template>
          <template #status-cell="{ row }">
            <UBadge
              :color="statusColor[row.original.status]"
              :icon="statusIcon[row.original.status]"
              variant="subtle"
              :label="statusLabel[row.original.status]"
            />
          </template>
          <template #error-cell="{ row }">
            <span class="text-sm text-muted">{{ row.original.error ?? '-' }}</span>
          </template>
          <template #processedAt-cell="{ row }"><TimeAgo :date="row.original.processedAt" /></template>
        </UTable>
      </template>

      <UModal v-model:open="cancelOpen" title="Cancelar remoção">
        <template #body>
          <p>
            As <strong>{{ pending }}</strong> remoção(ões) pendente(s) serão canceladas.
            Se alguém estiver sendo removido neste exato momento, essa remoção termina normalmente.
          </p>
        </template>
        <template #footer>
          <UButton label="Voltar" variant="ghost" color="neutral" @click="cancelOpen = false" />
          <UButton label="Cancelar pendentes" color="error" :loading="canceling" @click="cancel" />
        </template>
      </UModal>
    </template>
  </UDashboardPanel>
</template>

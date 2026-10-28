<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const { user, clear } = useUserSession()
const toast = useToast()
const { data: groups, refresh: refreshGroups } = await useGroups()
const { data: active, refresh: refreshActive } = await useActiveRemovals()

const pendingByGroup = computed(() => {
  const map = new Map<string, number>()
  for (const r of active.value) {
    map.set(r.groupId, (map.get(r.groupId) ?? 0) + r.pending)
  }
  return map
})

const items = computed<NavigationMenuItem[][]>(() => {
  const managed = groups.value.filter(g => g.managed)
  return [
    [
      { label: 'Todos os grupos', icon: 'i-lucide-layout-grid', to: '/', exact: true },
      { label: 'Buscar pessoa', icon: 'i-lucide-search', to: '/people' }
    ],
    managed.length
      ? [
          { label: 'Gerenciados', type: 'label' },
          ...managed.map((g) => {
            const pending = pendingByGroup.value.get(g.id)
            return {
              label: g.name,
              icon: pending ? 'i-lucide-loader' : 'i-lucide-message-circle',
              to: `/groups/${g.id}`,
              badge: pending ? { label: String(pending), color: 'warning' as const, variant: 'subtle' as const } : undefined,
              tooltip: pending ? { text: `${pending} remoção(ões) em andamento` } : undefined
            }
          })
        ]
      : []
  ]
})

// polls the running removals: feeds the sidebar badges and notices when one finishes
const known = ref(new Set(active.value.map(r => r.id)))
let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  timer = setInterval(async () => {
    await refreshActive()
    const now = new Set(active.value.map(r => r.id))
    for (const id of known.value) {
      if (!now.has(id)) await notifyFinished(id)
    }
    known.value = now
  }, 10_000)
})
onUnmounted(() => clearInterval(timer))

async function notifyFinished(id: number) {
  try {
    const job = await $fetch<RemovalDetail>(`/api/go/removals/${id}`)
    const ok = job.counts.ok ?? 0
    const failed = job.counts.failed ?? 0
    const canceled = job.counts.canceled ?? 0
    const parts = [`${ok} removido(s)`]
    if (failed) parts.push(`${failed} falha(s)`)
    if (canceled) parts.push(`${canceled} cancelada(s)`)

    toast.add({
      title: job.closeGroup && !canceled ? `Grupo encerrado (remoção #${id})` : `Remoção #${id} concluída`,
      description: parts.join(', '),
      color: failed ? 'warning' : 'success',
      icon: failed ? 'i-lucide-triangle-alert' : 'i-lucide-check',
      duration: 10_000,
      actions: [{ label: 'Ver', variant: 'outline', onClick: () => { navigateTo(`/removals/${id}`) } }]
    })
    // a closing changes the group list (closed groups stop being managed)
    if (job.closeGroup) await refreshGroups()
  } catch {
    // the notice is a nice-to-have, a failed fetch is not worth an error
  }
}

async function logout() {
  await clear()
  await navigateTo('/login')
}
</script>

<template>
  <UDashboardGroup>
    <UDashboardSidebar collapsible resizable :ui="{ footer: 'border-t border-default' }">
      <template #header="{ collapsed }">
        <NuxtLink to="/" class="flex items-center gap-2 font-semibold min-w-0">
          <UIcon name="i-lucide-users" class="size-5 shrink-0 text-primary" />
          <span v-if="!collapsed" class="truncate">Gestão de grupos</span>
        </NuxtLink>
      </template>

      <template #default="{ collapsed }">
        <UNavigationMenu
          :items="items"
          orientation="vertical"
          :collapsed="collapsed"
          tooltip
          :ui="{ linkLabel: 'truncate' }"
        />
      </template>

      <template #footer="{ collapsed }">
        <div class="flex w-full items-center gap-1" :class="collapsed ? 'flex-col' : ''">
          <UColorModeButton />
          <span v-if="!collapsed" class="flex-1 truncate text-sm text-muted">{{ user?.email }}</span>
          <UTooltip text="Sair">
            <UButton icon="i-lucide-log-out" color="neutral" variant="ghost" aria-label="Sair" @click="logout" />
          </UTooltip>
        </div>
      </template>
    </UDashboardSidebar>

    <slot />
  </UDashboardGroup>
</template>

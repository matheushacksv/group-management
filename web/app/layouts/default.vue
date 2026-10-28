<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const { user, clear } = useUserSession()
const { data: groups } = await useGroups()

const items = computed<NavigationMenuItem[][]>(() => {
  const managed = groups.value.filter(g => g.managed)
  return [
    [{ label: 'Todos os grupos', icon: 'i-lucide-layout-grid', to: '/', exact: true }],
    managed.length
      ? [
          { label: 'Gerenciados', type: 'label' },
          ...managed.map(g => ({ label: g.name, icon: 'i-lucide-message-circle', to: `/groups/${g.id}` }))
        ]
      : []
  ]
})

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

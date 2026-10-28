<script setup lang="ts">
// search runs while typing, waits a 300ms pause to not fire a request per key
const q = ref('')
const debounced = ref('')
let timer: ReturnType<typeof setTimeout> | undefined
watch(q, (v) => {
  clearTimeout(timer)
  timer = setTimeout(() => { debounced.value = v.trim() }, 300)
})

const people = ref<Person[]>([])
const status = ref<'idle' | 'pending' | 'success' | 'error'>('idle')
let last = 0
watch(debounced, async (v) => {
  if (v.length < 2) {
    people.value = []
    status.value = 'idle'
    return
  }
  // a slow answer for an old search must not overwrite the newer one
  const current = ++last
  status.value = 'pending'
  try {
    const res = await $fetch<Person[]>('/api/go/members/search', { query: { q: v } })
    if (current !== last) return
    people.value = res
    status.value = 'success'
  } catch {
    if (current !== last) return
    people.value = []
    status.value = 'error'
  }
})
</script>

<template>
  <UDashboardPanel id="people">
    <template #header>
      <UDashboardNavbar title="Buscar pessoa" />
      <UDashboardToolbar>
        <template #left>
          <UInput
            v-model="q"
            icon="i-lucide-search"
            placeholder="Número ou nome (mín. 2 caracteres)"
            class="w-80"
            autofocus
            :loading="status === 'pending'"
          />
        </template>
      </UDashboardToolbar>
    </template>

    <template #body>
      <UEmpty
        v-if="debounced.length < 2"
        icon="i-lucide-search"
        title="Busque alguém"
        description="Veja em quais grupos a pessoa está e de quais já saiu."
      />
      <UAlert v-else-if="status === 'error'" color="error" variant="subtle" title="Falha na busca, tente de novo" />
      <UEmpty
        v-else-if="status === 'success' && !people.length"
        icon="i-lucide-user-x"
        title="Ninguém encontrado"
        :description="`Nenhum membro com &quot;${debounced}&quot; no número ou no nome.`"
      />

      <div v-else class="grid gap-4 lg:grid-cols-2">
        <UCard v-for="p in people" :key="p.id">
          <template #header>
            <MemberCell :name="p.pushName" :number="p.number" />
          </template>
          <ul class="space-y-2">
            <li v-for="g in p.groups" :key="g.groupId" class="flex items-center justify-between gap-3">
              <NuxtLink :to="`/groups/${g.groupId}`" class="truncate hover:underline" :class="g.present ? 'text-highlighted' : 'text-muted'">
                {{ g.name }}
              </NuxtLink>
              <UBadge
                :color="g.present ? 'success' : 'neutral'"
                variant="subtle"
                class="shrink-0"
              >
                {{ g.present ? 'no grupo' : 'saiu' }} {{ timeAgo(g.since) }}
              </UBadge>
            </li>
          </ul>
        </UCard>
      </div>
    </template>
  </UDashboardPanel>
</template>

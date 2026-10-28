<script setup lang="ts">
const props = defineProps<{ group: StoredGroup }>()
const open = defineModel<boolean>('open', { default: false })
const toast = useToast()
const { refresh: refreshGroups } = useGroups()

// live participants from WhatsApp, to show exactly who goes and who stays
const live = ref<Group>()
const loading = ref(false)
const confirmText = ref('')
const closing = ref(false)

watch(open, async (isOpen) => {
  if (!isOpen) return
  confirmText.value = ''
  live.value = undefined
  loading.value = true
  try {
    live.value = await $fetch<Group>(groupsPath(props.group.id))
  } catch (err) {
    toast.add({ title: 'Falha ao ler o grupo no WhatsApp', description: apiError(err), color: 'error' })
    open.value = false
  } finally {
    loading.value = false
  }
})

const participants = computed(() => live.value?.participants ?? [])
const creators = computed(() => participants.value.filter(p => p.isSuperAdmin))
// creator(s) can't be removed and the bot leaves by itself at the end
const toRemove = computed(() => Math.max(participants.value.length - creators.value.length - 1, 0))
const confirmed = computed(() => confirmText.value.trim() === props.group.name.trim())

async function close() {
  closing.value = true
  try {
    const res = await $fetch<CloseGroupResponse>(`${groupsPath(props.group.id)}/close`, { method: 'POST' })
    open.value = false
    if (res.jobId) {
      toast.add({ title: 'Encerramento iniciado', description: `${res.queued} pessoa(s) na fila de remoção`, color: 'info', icon: 'i-lucide-loader' })
      await navigateTo(`/removals/${res.jobId}?group=${props.group.id}`)
    } else {
      toast.add({ title: 'Grupo encerrado', description: 'Não havia ninguém para remover, o bot saiu do grupo.', color: 'success', icon: 'i-lucide-check' })
      await refreshGroups()
      await navigateTo('/')
    }
  } catch (err) {
    toast.add({ title: 'Falha ao encerrar', description: apiError(err), color: 'error' })
  } finally {
    closing.value = false
  }
}
</script>

<template>
  <UModal v-model:open="open" title="Encerrar grupo" :description="group.name">
    <template #body>
      <div v-if="loading" class="flex items-center gap-2 text-muted">
        <UIcon name="i-lucide-loader" class="animate-spin" /> Lendo participantes no WhatsApp...
      </div>

      <div v-else-if="live" class="space-y-4">
        <UAlert
          color="error"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          title="Ação irreversível"
          description="Quem for removido pode não conseguir voltar ao grupo."
        />

        <ul class="space-y-2 text-sm">
          <li class="flex gap-2">
            <UIcon name="i-lucide-user-minus" class="size-4 mt-0.5 text-error shrink-0" />
            <span><strong>{{ toRemove }}</strong> pessoa(s) serão removidas, <strong>inclusive admins</strong>, na fila lenta (1 a cada 3–6s). Dá para cancelar no meio.</span>
          </li>
          <li class="flex gap-2">
            <UIcon name="i-lucide-log-out" class="size-4 mt-0.5 text-warning shrink-0" />
            <span>Ao terminar <strong>sem cancelamento</strong>, o bot sai do grupo e ele some do painel.</span>
          </li>
          <li v-if="creators.length" class="flex gap-2">
            <UIcon name="i-lucide-crown" class="size-4 mt-0.5 text-muted shrink-0" />
            <span>
              O criador ({{ creators.map(c => c.number).join(', ') }}) <strong>não pode ser removido</strong> pelo WhatsApp:
              para o grupo sumir de vez, ele precisa tocar em <em>Excluir grupo</em> no celular.
            </span>
          </li>
        </ul>

        <UFormField label="Digite o nome do grupo para confirmar">
          <UInput v-model="confirmText" :placeholder="group.name" class="w-full" />
        </UFormField>
      </div>
    </template>

    <template #footer>
      <UButton label="Voltar" variant="ghost" color="neutral" @click="open = false" />
      <UButton
        label="Encerrar grupo"
        color="error"
        icon="i-lucide-power"
        :loading="closing"
        :disabled="!confirmed || loading"
        @click="close"
      />
    </template>
  </UModal>
</template>

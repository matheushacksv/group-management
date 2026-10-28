<script setup lang="ts">
const props = defineProps<{ group: StoredGroup }>()
const open = defineModel<boolean>('open', { default: false })
const toast = useToast()
const { refresh: refreshGroups } = useGroups()

const confirmText = ref('')
const deleting = ref(false)
watch(open, (isOpen) => {
  if (isOpen) confirmText.value = ''
})

const confirmed = computed(() => confirmText.value.trim() === props.group.name.trim())

async function remove() {
  deleting.value = true
  try {
    await $fetch(groupsPath(props.group.id), { method: 'DELETE' })
    open.value = false
    toast.add({ title: 'Grupo excluído', description: props.group.name, color: 'success', icon: 'i-lucide-check' })
    await refreshGroups()
    if (useRoute().path !== '/') await navigateTo('/')
  } catch (err) {
    toast.add({ title: 'Falha ao excluir', description: apiError(err), color: 'error' })
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <UModal v-model:open="open" title="Excluir grupo" :description="group.name">
    <template #body>
      <div class="space-y-4">
        <UAlert
          color="error"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          title="Apaga todos os dados do grupo no painel"
          description="Histórico de entradas e saídas, remoções e membros que não estão em nenhum outro grupo. Não dá para desfazer. O WhatsApp não é alterado: o bot já saiu do grupo quando ele foi encerrado."
        />
        <UFormField label="Digite o nome do grupo para confirmar">
          <UInput v-model="confirmText" :placeholder="group.name" class="w-full" />
        </UFormField>
      </div>
    </template>

    <template #footer>
      <UButton label="Voltar" variant="ghost" color="neutral" @click="open = false" />
      <UButton
        label="Excluir grupo"
        color="error"
        icon="i-lucide-trash-2"
        :loading="deleting"
        :disabled="!confirmed"
        @click="remove"
      />
    </template>
  </UModal>
</template>

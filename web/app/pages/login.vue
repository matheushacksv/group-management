<script setup lang="ts">
definePageMeta({ layout: false })

const { fetch: refreshSession } = useUserSession()
const toast = useToast()

const state = reactive({ email: '', password: '' })
const loading = ref(false)

async function onSubmit() {
  loading.value = true
  try {
    await $fetch('/api/login', { method: 'POST', body: state })
    await refreshSession()
    await navigateTo('/')
  } catch {
    toast.add({ title: 'Email ou senha inválidos', color: 'error', icon: 'i-lucide-circle-x' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-dvh flex items-center justify-center p-4 bg-muted">
    <UCard class="w-full max-w-sm">
      <template #header>
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-users" class="size-5 text-primary" />
          <h1 class="font-semibold">Gestão de grupos</h1>
        </div>
      </template>

      <UForm :state="state" class="space-y-4" @submit="onSubmit">
        <UFormField name="email" label="Email" required>
          <UInput v-model="state.email" type="email" autocomplete="username" class="w-full" />
        </UFormField>
        <UFormField name="password" label="Senha" required>
          <UInput v-model="state.password" type="password" autocomplete="current-password" class="w-full" />
        </UFormField>
        <UButton type="submit" label="Entrar" block :loading="loading" />
      </UForm>
    </UCard>
  </div>
</template>

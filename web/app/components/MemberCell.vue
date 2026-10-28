<script setup lang="ts">
defineProps<{ name?: string, number: string }>()
</script>

<template>
  <div class="flex items-center gap-3 min-w-0">
    <!-- alt generates the initials; without a name falls back to a person icon -->
    <UAvatar :alt="name" icon="i-lucide-user" size="sm" />
    <div class="min-w-0">
      <p class="truncate font-medium text-highlighted">{{ name ?? number }}</p>
      <p v-if="name" class="truncate text-xs text-muted">{{ number }}</p>
    </div>
  </div>
</template>

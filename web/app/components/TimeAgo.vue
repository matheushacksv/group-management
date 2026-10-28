<script setup lang="ts">
defineProps<{ date?: string }>()
</script>

<template>
  <UTooltip v-if="date" :text="formatDate(date)">
    <span>{{ timeAgo(date) }}</span>
  </UTooltip>
  <span v-else class="text-muted">-</span>
</template>

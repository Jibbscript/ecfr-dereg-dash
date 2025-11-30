<template>
  <UsaAccordion v-if="title">
    <UsaAccordionItem title="Metrics">
      <p>Word Count: {{ title.total_words }}</p>
      <p>RSCS: {{ title.avg_rscs }}</p>
    </UsaAccordionItem>
    <UsaAccordionItem title="LSA Counts">
      <!-- Counts -->
    </UsaAccordionItem>
    <UsaAccordionItem title="Summary">
      {{ title.summary }}
    </UsaAccordionItem>
  </UsaAccordion>
  <div v-else-if="error">
    Error loading title.
  </div>
  <div v-else>
    Loading...
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'

const route = useRoute()
const { data: title, error } = await useFetch(`/api/titles/${route.params.t}`)

if (error.value) {
  console.error('Error loading title:', error.value)
}
</script>

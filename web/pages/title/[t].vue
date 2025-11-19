<template>
  <UsaAccordion>
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
</template>

<script setup>
import { useRoute } from 'vue-router'

const route = useRoute()
const { data: title, error } = await useFetch(`/api/titles/${route.params.t}`)

if (error.value) {
  console.error('Error loading title:', error.value)
}
</script>

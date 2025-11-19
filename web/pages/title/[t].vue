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
import { ref } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const title = ref({})

const fetchTitle = async () => {
  const res = await fetch(`/api/titles/${route.params.t}`)
  title.value = await res.json()
}

fetchTitle()
</script>

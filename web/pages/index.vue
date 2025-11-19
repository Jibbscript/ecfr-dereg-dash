<template>
  <ClientOnly>
    <div v-if="loading">Loading...</div>
    <div v-else-if="error" class="usa-alert usa-alert--error">
      <div class="usa-alert__body">
        <h4 class="usa-alert__heading">Error</h4>
        <p class="usa-alert__text">{{ error }}</p>
      </div>
    </div>
    <table v-else class="usa-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Word Count</th>
          <th>RSCS per 1K</th>
          <th>LSA Counts</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="agency in agencies" :key="agency.id">
          <td>{{ agency.name }}</td>
          <td>{{ agency.total_words || 0 }}</td>
          <td>{{ agency.avg_rscs || 0 }}</td>
          <td>{{ agency.lsa_counts || 0 }}</td>
        </tr>
      </tbody>
    </table>
  </ClientOnly>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const columns = [
  { id: 'name', label: 'Name' },
  { id: 'total_words', label: 'Word Count', sortable: true },
  { id: 'avg_rscs', label: 'RSCS per 1K', sortable: true },
  { id: 'lsa_counts', label: 'LSA Counts', sortable: true }
]

const agencies = ref([])
const loading = ref(true)
const error = ref(null)

onMounted(async () => {
  try {
    console.log('Fetching agencies...')
    const res = await fetch('/api/agencies')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    agencies.value = await res.json()
    console.log('Agencies loaded:', agencies.value)
  } catch (e) {
    console.error('Fetch error:', e)
    error.value = e.message
  } finally {
    loading.value = false
  }
})

function handleSort({ column, dir }) {
  // Fetch with sort params
}
</script>

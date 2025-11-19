<template>
  <ClientOnly>
    <table class="usa-table">
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

onMounted(async () => {
  try {
    console.log('Fetching agencies...')
    const res = await fetch('/api/agencies')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    agencies.value = await res.json()
    console.log('Agencies loaded:', agencies.value)
  } catch (e) {
    console.error('Fetch error:', e)
  }
})

function handleSort({ column, dir }) {
  // Fetch with sort params
}
</script>

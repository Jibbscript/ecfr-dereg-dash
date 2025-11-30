<template>
  <div class="grid-container">
    <!-- Summary Card -->
    <section class="usa-summary-box margin-bottom-4" aria-labelledby="summary-heading">
      <div class="usa-summary-box__body">
        <h2 class="usa-summary-box__heading" id="summary-heading">Regulatory Overview</h2>
        <div class="usa-summary-box__text">
          <p v-if="!loading && !error">
            {{ totalWords.toLocaleString() }} words across {{ parentAgencies.length }} departments.
            Average RSCS: {{ avgRSCS.toFixed(1) }} per 1,000 words.
          </p>
          <p v-else-if="loading">Loading data...</p>
        </div>
      </div>
    </section>

    <!-- Title Filter -->
    <div class="usa-form-group margin-bottom-3">
      <label class="usa-label" for="title-filter">Filter by CFR Title</label>
      <select id="title-filter" class="usa-select" v-model="selectedTitle" @change="fetchAgencies">
        <option value="">All Titles</option>
        <option v-for="t in 50" :key="t" :value="String(t)">Title {{ t }}</option>
      </select>
    </div>

    <!-- Checksum Toggle -->
    <div class="usa-checkbox margin-bottom-3">
      <input
        class="usa-checkbox__input"
        id="include-checksum"
        type="checkbox"
        v-model="includeChecksum"
        @change="fetchAgencies"
      />
      <label class="usa-checkbox__label" for="include-checksum">
        Show Content Checksums
      </label>
    </div>

    <!-- Loading/Error States -->
    <div v-if="loading" class="usa-alert usa-alert--info">
      <div class="usa-alert__body">
        <p class="usa-alert__text">Loading agency data...</p>
      </div>
    </div>
    <div v-else-if="error" class="usa-alert usa-alert--error">
      <div class="usa-alert__body">
        <h4 class="usa-alert__heading">Error</h4>
        <p class="usa-alert__text">{{ error }}</p>
      </div>
    </div>

    <!-- Hierarchical Data Table -->
    <table v-else class="usa-table usa-table--borderless" style="width: 100%;">
      <thead>
        <tr>
          <th scope="col" @click="sortBy('name')" style="cursor: pointer;">
            Agency Name
            <span v-if="sortKey === 'name'">{{ sortDir === 'asc' ? ' ▲' : ' ▼' }}</span>
          </th>
          <th scope="col" @click="sortBy('total_words')" style="cursor: pointer;">
            Word Count
            <span v-if="sortKey === 'total_words'">{{ sortDir === 'asc' ? ' ▲' : ' ▼' }}</span>
          </th>
          <th scope="col" @click="sortBy('avg_rscs')" style="cursor: pointer;">
            RSCS per 1K
            <span v-if="sortKey === 'avg_rscs'">{{ sortDir === 'asc' ? ' ▲' : ' ▼' }}</span>
          </th>
          <th scope="col">LSA Counts</th>
          <th v-if="includeChecksum" scope="col">Content Checksum</th>
        </tr>
      </thead>
      <tbody>
        <template v-for="parent in sortedParents" :key="parent.id">
          <!-- Parent row (department) -->
          <tr
            class="parent-row"
            :style="{ cursor: getChildren(parent.id).length ? 'pointer' : 'default', backgroundColor: '#f0f0f0' }"
            @click="toggleExpand(parent.id)"
          >
            <th scope="row">
              <span v-if="getChildren(parent.id).length" style="margin-right: 0.5rem;">
                {{ expanded[parent.id] ? '▼' : '▶' }}
              </span>
              <strong>{{ parent.name }}</strong>
            </th>
            <td>{{ (parent.total_words || 0).toLocaleString() }}</td>
            <td>{{ (parent.avg_rscs || 0).toFixed(1) }}</td>
            <td>{{ parent.lsa_counts || 0 }}</td>
            <td v-if="includeChecksum" class="checksum-cell">
              <code v-if="parent.content_checksum" :title="parent.content_checksum">
                {{ truncateChecksum(parent.content_checksum) }}
              </code>
              <span v-else>-</span>
            </td>
          </tr>
          <!-- Child rows (sub-agencies) -->
          <template v-if="expanded[parent.id]">
            <tr v-for="child in getChildren(parent.id)" :key="child.id" class="child-row">
              <td style="padding-left: 2rem;">↳ {{ child.name }}</td>
              <td>{{ (child.total_words || 0).toLocaleString() }}</td>
              <td>{{ (child.avg_rscs || 0).toFixed(1) }}</td>
              <td>{{ child.lsa_counts || 0 }}</td>
              <td v-if="includeChecksum" class="checksum-cell">
                <code v-if="child.content_checksum" :title="child.content_checksum">
                  {{ truncateChecksum(child.content_checksum) }}
                </code>
                <span v-else>-</span>
              </td>
            </tr>
          </template>
        </template>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const selectedTitle = ref('')
const includeChecksum = ref(false)
const agencies = ref([])
const loading = ref(true)
const error = ref(null)
const expanded = ref({})
const sortKey = ref('total_words')
const sortDir = ref('desc')

// Truncate checksum for display (show first 8 chars)
function truncateChecksum(hash) {
  if (!hash) return ''
  return hash.substring(0, 8) + '...'
}

// Separate parents (departments) from children (sub-agencies)
const parentAgencies = computed(() =>
  agencies.value.filter(a => !a.parent_id)
)

const sortedParents = computed(() => {
  return [...parentAgencies.value].sort((a, b) => {
    let aVal = a[sortKey.value]
    let bVal = b[sortKey.value]

    // Handle string sorting for name
    if (sortKey.value === 'name') {
      aVal = (aVal || '').toLowerCase()
      bVal = (bVal || '').toLowerCase()
      if (sortDir.value === 'asc') {
        return aVal.localeCompare(bVal)
      }
      return bVal.localeCompare(aVal)
    }

    // Numeric sorting
    aVal = aVal || 0
    bVal = bVal || 0
    return sortDir.value === 'asc' ? aVal - bVal : bVal - aVal
  })
})

function getChildren(parentId) {
  return agencies.value.filter(a => a.parent_id === parentId)
}

function toggleExpand(id) {
  expanded.value[id] = !expanded.value[id]
}

const totalWords = computed(() =>
  parentAgencies.value.reduce((sum, a) => sum + (a.total_words || 0), 0)
)

const avgRSCS = computed(() => {
  const vals = parentAgencies.value.filter(a => a.avg_rscs > 0)
  if (!vals.length) return 0
  return vals.reduce((s, a) => s + a.avg_rscs, 0) / vals.length
})

async function fetchAgencies() {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams()
    if (selectedTitle.value) {
      params.set('title', selectedTitle.value)
    }
    if (includeChecksum.value) {
      params.set('include_checksum', 'true')
    }
    const qp = params.toString() ? `?${params.toString()}` : ''
    const res = await fetch(`/api/agencies${qp}`)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    agencies.value = data || []
  } catch (e) {
    console.error('Fetch error:', e)
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function sortBy(key) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'desc'
  }
}

onMounted(fetchAgencies)
</script>

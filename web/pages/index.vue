<template>
  <div>
    <!-- Hero Section -->
    <section class="usa-hero" aria-label="Introduction">
      <div class="grid-container">
        <div class="usa-hero__callout">
          <h1 class="usa-hero__heading">
            <span class="usa-hero__heading--alt">Federal Regulatory</span>
            Transparency Dashboard
          </h1>
          <p>
            Explore and analyze the complexity of federal regulations across all agencies. 
            Track regulatory burden using the RSCS metric and identify opportunities for simplification.
          </p>
          <UsaButton variant="secondary" @click="showRscsModal = true">
            Learn about RSCS
          </UsaButton>
        </div>
      </div>
    </section>

    <!-- Summary Metrics Cards -->
    <section class="grid-container margin-top-4" aria-label="Summary statistics">
      <div class="grid-row grid-gap">
        <div class="tablet:grid-col-3">
          <MetricCard
            title="Total Regulatory Words"
            :value="totalWords"
            description="Across all tracked agencies"
          />
        </div>
        <div class="tablet:grid-col-3">
          <MetricCard
            title="Federal Agencies"
            :value="parentAgencies.length"
            description="Departments & sub-agencies"
          />
        </div>
        <div class="tablet:grid-col-3">
          <MetricCard
            title="Avg. RSCS Score"
            :value="avgRSCS"
            format="decimal"
            description="Per 1,000 words"
            :has-info="true"
            @info-click="showRscsModal = true"
          />
        </div>
        <div class="tablet:grid-col-3">
          <MetricCard
            title="LSA Activity"
            :value="totalLSA"
            description="Recent regulatory changes"
          />
        </div>
      </div>
    </section>

    <!-- RSCS Explainer Modal -->
    <RscsExplainer :visible="showRscsModal" @close="showRscsModal = false" />

    <!-- Data Section -->
    <section class="grid-container margin-top-4" aria-label="Agency data">
      <h2 class="font-heading-xl margin-bottom-2">Agency Regulatory Metrics</h2>

      <!-- Filters -->
      <div class="grid-row grid-gap margin-bottom-3">
        <div class="tablet:grid-col-4">
          <div class="usa-form-group">
            <label class="usa-label" for="title-filter">Filter by CFR Title</label>
            <select id="title-filter" class="usa-select" v-model="selectedTitle" @change="fetchAgencies">
              <option value="">All Titles</option>
              <option v-for="t in 50" :key="t" :value="String(t)">Title {{ t }}</option>
            </select>
          </div>
        </div>
        <div class="tablet:grid-col-4 display-flex flex-align-end">
          <div class="usa-checkbox">
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
        </div>
      </div>

      <!-- Loading State -->
      <LoadingSkeleton v-if="loading" variant="table" :count="8" :columns="5" />

      <!-- Error State -->
      <UsaAlert v-else-if="error" status="error" class="margin-bottom-2">
        <template #heading>Error Loading Data</template>
        {{ error }}
      </UsaAlert>

      <!-- Data Table -->
      <div v-else class="usa-table-container--scrollable" tabindex="0">
        <table class="usa-table usa-table--striped" style="width: 100%;">
          <thead>
            <tr>
              <th scope="col">
                <button
                  type="button"
                  class="usa-button--unstyled sortable-header"
                  @click="sortBy('name')"
                  :aria-sort="sortKey === 'name' ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'"
                >
                  Agency Name
                  <span class="sort-indicator" v-if="sortKey === 'name'">
                    {{ sortDir === 'asc' ? '▲' : '▼' }}
                  </span>
                </button>
              </th>
              <th scope="col">
                <button
                  type="button"
                  class="usa-button--unstyled sortable-header"
                  @click="sortBy('total_words')"
                  :aria-sort="sortKey === 'total_words' ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'"
                >
                  Word Count
                  <span class="sort-indicator" v-if="sortKey === 'total_words'">
                    {{ sortDir === 'asc' ? '▲' : '▼' }}
                  </span>
                </button>
              </th>
              <th scope="col">
                <button
                  type="button"
                  class="usa-button--unstyled sortable-header"
                  @click="sortBy('avg_rscs')"
                  :aria-sort="sortKey === 'avg_rscs' ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'"
                >
                  RSCS per 1K
                  <button
                    type="button"
                    class="usa-button--unstyled info-icon"
                    aria-label="Information about RSCS metric"
                    @click.stop="showRscsModal = true"
                  >
                    ⓘ
                  </button>
                  <span class="sort-indicator" v-if="sortKey === 'avg_rscs'">
                    {{ sortDir === 'asc' ? '▲' : '▼' }}
                  </span>
                </button>
              </th>
              <th scope="col">LSA Activity</th>
              <th v-if="includeChecksum" scope="col">Content Checksum</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="parent in sortedParents" :key="parent.id">
              <!-- Parent row (department) -->
              <tr
                class="parent-row"
                :class="{ 'has-children': getChildren(parent.id).length > 0 }"
                @click="toggleExpand(parent.id)"
                @keydown.enter="toggleExpand(parent.id)"
                :tabindex="getChildren(parent.id).length ? 0 : -1"
                :aria-expanded="getChildren(parent.id).length ? expanded[parent.id] : undefined"
              >
                <th scope="row">
                  <span v-if="getChildren(parent.id).length" class="expand-icon" aria-hidden="true">
                    {{ expanded[parent.id] ? '▼' : '▶' }}
                  </span>
                  <strong>{{ parent.name }}</strong>
                </th>
                <td>{{ (parent.total_words || 0).toLocaleString() }}</td>
                <td>
                  <span :class="getRscsClass(parent.avg_rscs)">
                    {{ (parent.avg_rscs || 0).toFixed(1) }}
                  </span>
                </td>
                <td>{{ parent.lsa_counts || 0 }}</td>
                <td v-if="includeChecksum" class="checksum-cell">
                  <code v-if="parent.content_checksum" :title="parent.content_checksum">
                    {{ truncateChecksum(parent.content_checksum) }}
                  </code>
                  <span v-else class="text-base">—</span>
                </td>
              </tr>
              <!-- Child rows (sub-agencies) -->
              <template v-if="expanded[parent.id]">
                <tr v-for="child in getChildren(parent.id)" :key="child.id" class="child-row">
                  <td class="padding-left-4">
                    <span class="text-base-dark">↳</span> {{ child.name }}
                  </td>
                  <td>{{ (child.total_words || 0).toLocaleString() }}</td>
                  <td>
                    <span :class="getRscsClass(child.avg_rscs)">
                      {{ (child.avg_rscs || 0).toFixed(1) }}
                    </span>
                  </td>
                  <td>{{ child.lsa_counts || 0 }}</td>
                  <td v-if="includeChecksum" class="checksum-cell">
                    <code v-if="child.content_checksum" :title="child.content_checksum">
                      {{ truncateChecksum(child.content_checksum) }}
                    </code>
                    <span v-else class="text-base">—</span>
                  </td>
                </tr>
              </template>
            </template>
          </tbody>
        </table>
      </div>

      <!-- Empty State -->
      <div v-if="!loading && !error && sortedParents.length === 0" class="text-center padding-4">
        <p class="text-base">No agency data available for the selected filter.</p>
      </div>
    </section>
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
const showRscsModal = ref(false)

// Truncate checksum for display (show first 8 chars)
function truncateChecksum(hash) {
  if (!hash) return ''
  return hash.substring(0, 8) + '...'
}

// Get RSCS color class based on value
function getRscsClass(rscs) {
  if (!rscs) return ''
  if (rscs < 1050) return 'rscs-low'
  if (rscs < 1150) return 'rscs-medium'
  return 'rscs-high'
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
  if (getChildren(id).length > 0) {
    expanded.value[id] = !expanded.value[id]
  }
}

const totalWords = computed(() =>
  parentAgencies.value.reduce((sum, a) => sum + (a.total_words || 0), 0)
)

const avgRSCS = computed(() => {
  const vals = parentAgencies.value.filter(a => a.avg_rscs > 0)
  if (!vals.length) return 0
  return vals.reduce((s, a) => s + a.avg_rscs, 0) / vals.length
})

const totalLSA = computed(() =>
  parentAgencies.value.reduce((sum, a) => sum + (a.lsa_counts || 0), 0)
)

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

<style scoped>
.usa-hero {
  background-color: #1a4480;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100' viewBox='0 0 100 100'%3E%3Cg fill-rule='evenodd'%3E%3Cg fill='%23162e51' fill-opacity='0.4'%3E%3Cpath opacity='.5' d='M96 95h4v1h-4v4h-1v-4h-9v4h-1v-4h-9v4h-1v-4h-9v4h-1v-4h-9v4h-1v-4h-9v4h-1v-4h-9v4h-1v-4h-9v4h-1v-4h-9v4h-1v-4H0v-1h15v-9H0v-1h15v-9H0v-1h15v-9H0v-1h15v-9H0v-1h15v-9H0v-1h15v-9H0v-1h15v-9H0v-1h15v-9H0v-1h15V0h1v15h9V0h1v15h9V0h1v15h9V0h1v15h9V0h1v15h9V0h1v15h9V0h1v15h9V0h1v15h9V0h1v15h4v1h-4v9h4v1h-4v9h4v1h-4v9h4v1h-4v9h4v1h-4v9h4v1h-4v9h4v1h-4v9h4v1h-4v9zm-1 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-9-10h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm9-10v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-9-10h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm9-10v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-9-10h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm9-10v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-10 0v-9h-9v9h9zm-9-10h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9zm10 0h9v-9h-9v9z'/%3E%3Cpath d='M6 5V0H5v5H0v1h5v94h1V6h94V5H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
  padding: 2rem 0 3rem;
}

.usa-hero__callout {
  background-color: rgba(26, 68, 128, 0.9);
  max-width: 36rem;
  padding: 2rem;
}

.usa-hero__heading {
  color: #fff;
  font-size: 2.5rem;
  line-height: 1.2;
}

.usa-hero__heading--alt {
  display: block;
  font-size: 1.5rem;
  color: #a9aeb1;
}

.usa-hero p {
  color: #fff;
  margin-bottom: 1.5rem;
}

.sortable-header {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  background: none;
  border: none;
  cursor: pointer;
  font-weight: 700;
  color: inherit;
  padding: 0;
  text-align: left;
}

.sortable-header:hover {
  text-decoration: underline;
}

.sort-indicator {
  font-size: 0.75rem;
  color: #005ea2;
}

.info-icon {
  font-size: 1rem;
  color: #005ea2;
  padding: 0 0.25rem;
  background: none;
  border: none;
  cursor: pointer;
  vertical-align: middle;
}

.info-icon:hover {
  color: #1a4480;
}

.parent-row.has-children {
  cursor: pointer;
}

.parent-row.has-children:hover {
  background-color: #f0f0f0;
}

.expand-icon {
  display: inline-block;
  width: 1rem;
  margin-right: 0.5rem;
  font-size: 0.75rem;
  color: #71767a;
}

.child-row {
  background-color: #f9f9f9;
}

.child-row td:first-child {
  padding-left: 2rem;
}

.checksum-cell code {
  font-size: 0.8rem;
  background-color: #f0f0f0;
  padding: 0.125rem 0.25rem;
  border-radius: 2px;
}

.rscs-low {
  color: #4d8055;
  font-weight: 600;
}

.rscs-medium {
  color: #c05600;
  font-weight: 600;
}

.rscs-high {
  color: #b50909;
  font-weight: 600;
}

.text-center {
  text-align: center;
}
</style>

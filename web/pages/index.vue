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
<UsaButton variant="secondary" @click="explainer.open($event && $event.currentTarget)" class="hero-btn">
            <svg class="usa-icon margin-right-05" aria-hidden="true" focusable="false" width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/>
            </svg>
            Learn about RSCS
          </UsaButton>
          <UsaButton @click="showAiSummaries = true" class="margin-left-2 hero-btn">
            <svg class="usa-icon margin-right-05" aria-hidden="true" focusable="false" width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/>
            </svg>
            AI Summaries
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
@info-click="explainer.open($event && $event.currentTarget)"
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

    <!-- RSCS Explainer Modal now provided globally in layout -->
    <AiSummariesModal v-model:visible="showAiSummaries" />

    <!-- Data Section -->
    <section class="grid-container margin-top-4" aria-label="Agency data">
      <div class="section-header">
        <svg class="section-icon" aria-hidden="true" width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
          <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/>
        </svg>
        <h2 class="font-heading-xl margin-bottom-0">Agency Regulatory Metrics</h2>
      </div>
      <p class="section-subtitle">Explore regulatory complexity data across all federal agencies</p>

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
                <div class="th-header-group">
                  <button
                    type="button"
                    class="usa-button--unstyled sortable-header"
                    @click="sortBy('avg_rscs')"
                    :aria-sort="sortKey === 'avg_rscs' ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'"
                    aria-label="Sort by RSCS per 1K"
                  >
                    RSCS per 1K
                    <span class="sort-indicator" v-if="sortKey === 'avg_rscs'">
                      {{ sortDir === 'asc' ? '▲' : '▼' }}
                    </span>
                  </button>
                  <button
                    type="button"
                    class="usa-button--unstyled info-icon"
                    aria-label="Open RSCS explanation"
                    aria-haspopup="dialog"
                    aria-controls="rscs-explainer"
                    @click.stop="explainer.open($event && $event.currentTarget)"
                  >
                    <svg class="usa-icon" aria-hidden="true" focusable="false" role="img" width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/>
                    </svg>
                  </button>
                </div>
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
import { useRscsExplainer } from '../composables/useRscsExplainer'
import AiSummariesModal from '../components/AiSummariesModal.vue'

const selectedTitle = ref('')
const includeChecksum = ref(false)
const agencies = ref([])
const loading = ref(true)
const error = ref(null)
const expanded = ref({})
const sortKey = ref('total_words')
const sortDir = ref('desc')
const explainer = useRscsExplainer()
const showAiSummaries = ref(false)

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
  background-image: url('/hero-bg.svg');
  background-size: cover;
  background-position: center;
  padding: 3rem 0 4rem;
  position: relative;
}

.usa-hero__callout {
  background-color: rgba(22, 46, 81, 0.95);
  max-width: 40rem;
  padding: 2.5rem;
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(8px);
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
  opacity: 0.95;
  line-height: 1.6;
}

.hero-btn {
  display: inline-flex;
  align-items: center;
}

.hero-btn .usa-icon {
  width: 1.25rem;
  height: 1.25rem;
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

.th-header-group {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.info-icon {
  color: #005ea2;
  padding: 0 0.25rem;
  background: none;
  border: none;
  cursor: pointer;
  vertical-align: middle;
  display: inline-flex;
  align-items: center;
  transition: color 0.15s ease;
}

.info-icon:hover {
  color: #1a4480;
}

.usa-icon {
  width: 1.125rem;
  height: 1.125rem;
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

/* Section Header */
.section-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.section-icon {
  color: #005ea2;
}

.section-subtitle {
  color: #71767a;
  font-size: 1rem;
  margin-top: 0;
  margin-bottom: 1.5rem;
}

/* Table Enhancements */
.usa-table-container--scrollable {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.usa-table {
  border-collapse: separate;
  border-spacing: 0;
}

.usa-table thead th {
  background-color: #f0f0f0;
  border-bottom: 2px solid #005ea2;
}

.usa-table tbody tr {
  transition: background-color 0.15s ease;
}
</style>

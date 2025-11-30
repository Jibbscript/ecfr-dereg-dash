<template>
  <div class="grid-container">
    <!-- Breadcrumbs -->
    <UsaBreadcrumb class="margin-top-2">
      <UsaBreadcrumbItem href="/">Dashboard</UsaBreadcrumbItem>
      <UsaBreadcrumbItem v-if="section" :href="`/title/${section.title || '1'}`">
        Title {{ section.title || '—' }}
      </UsaBreadcrumbItem>
      <UsaBreadcrumbItem :current="true">
        {{ section?.section || `Section ${route.params.id}` }}
      </UsaBreadcrumbItem>
    </UsaBreadcrumb>

    <!-- Loading State -->
    <LoadingSkeleton v-if="pending" variant="text" :count="8" />

    <!-- Error State -->
    <UsaAlert v-else-if="error" status="error" class="margin-top-2">
      <template #heading>Error Loading Section</template>
      Unable to load data for this section. Please try again later.
    </UsaAlert>

    <!-- Content -->
    <template v-else-if="section">
      <!-- Page Header -->
      <h1 class="font-heading-2xl margin-top-3 margin-bottom-1">
        {{ section.section }}
      </h1>
      <p class="text-base margin-top-0 margin-bottom-3">
        Part of Title {{ section.title || '—' }}
      </p>

      <!-- Summary Box -->
      <UsaSummaryBox class="margin-bottom-4">
        <template #heading>Section Overview</template>
        <div class="grid-row grid-gap">
          <div class="tablet:grid-col-6">
            <p class="margin-0">
              <strong>RSCS Score:</strong>
              <button
                type="button"
                class="usa-button--unstyled info-icon"
                aria-label="Information about RSCS metric"
                @click="showRscsModal = true"
              >
                ⓘ
              </button>
            </p>
            <p class="font-heading-xl margin-top-05 margin-bottom-0" :class="getRscsClass(section.rscs_per_1k)">
              {{ (section.rscs_per_1k || 0).toFixed(1) }}
              <span class="text-base font-body-sm">per 1,000 words</span>
            </p>
          </div>
          <div class="tablet:grid-col-6">
            <p class="margin-0"><strong>Revision Date:</strong></p>
            <p class="margin-top-05 margin-bottom-0">
              {{ section.rev_date || 'Not available' }}
            </p>
          </div>
        </div>
      </UsaSummaryBox>

      <!-- Metrics Grid -->
      <h2 class="font-heading-lg margin-bottom-2">Complexity Breakdown</h2>
      <div class="grid-row grid-gap margin-bottom-4">
        <div class="tablet:grid-col-3">
          <div class="metric-tile">
            <span class="metric-tile__value">{{ (section.word_count || 0).toLocaleString() }}</span>
            <span class="metric-tile__label">Words</span>
            <span class="metric-tile__weight">Weight: ×1</span>
          </div>
        </div>
        <div class="tablet:grid-col-3">
          <div class="metric-tile">
            <span class="metric-tile__value">{{ section.def_count || 0 }}</span>
            <span class="metric-tile__label">Definitions</span>
            <span class="metric-tile__weight">Weight: ×20</span>
          </div>
        </div>
        <div class="tablet:grid-col-3">
          <div class="metric-tile">
            <span class="metric-tile__value">{{ section.xref_count || 0 }}</span>
            <span class="metric-tile__label">Cross-References</span>
            <span class="metric-tile__weight">Weight: ×50</span>
          </div>
        </div>
        <div class="tablet:grid-col-3">
          <div class="metric-tile">
            <span class="metric-tile__value">{{ section.modal_count || 0 }}</span>
            <span class="metric-tile__label">Modal Verbs</span>
            <span class="metric-tile__weight">Weight: ×100</span>
          </div>
        </div>
      </div>

      <!-- Raw RSCS Calculation -->
      <div class="rscs-calculation margin-bottom-4">
        <h3 class="font-heading-sm margin-bottom-1">RSCS Calculation</h3>
        <p class="text-base margin-0">
          {{ section.word_count || 0 }} + 
          (20 × {{ section.def_count || 0 }}) + 
          (50 × {{ section.xref_count || 0 }}) + 
          (100 × {{ section.modal_count || 0 }}) = 
          <strong>{{ section.rscs_raw || 0 }}</strong> raw score
        </p>
      </div>

      <!-- AI Summary -->
      <div v-if="section.summary" class="margin-bottom-4">
        <h2 class="font-heading-lg margin-bottom-2">AI Summary</h2>
        <div class="summary-content">
          <p>{{ section.summary }}</p>
        </div>
      </div>

      <!-- Section Text -->
      <div class="margin-bottom-4">
        <h2 class="font-heading-lg margin-bottom-2">Regulatory Text</h2>
        <UsaAlert status="info" class="margin-bottom-2" slim>
          <template #heading>Excerpt</template>
          Showing the first 2,000 characters. 
          <a 
            :href="`https://www.ecfr.gov/current/title-${section.title}`" 
            target="_blank" 
            rel="noopener"
            class="usa-link"
          >
            View full text on eCFR.gov
          </a>
        </UsaAlert>
        <div class="section-text">
          <p>{{ truncatedText }}</p>
        </div>
      </div>

      <!-- External Link -->
      <div class="margin-bottom-4">
        <a 
          :href="`https://www.ecfr.gov/current/title-${section.title}`" 
          target="_blank" 
          rel="noopener"
          class="usa-button"
        >
          View on eCFR.gov
        </a>
      </div>

      <!-- Back Link -->
      <div class="margin-top-4 margin-bottom-4">
        <NuxtLink to="/" class="usa-link">
          ← Back to Dashboard
        </NuxtLink>
      </div>
    </template>

    <!-- RSCS Explainer Modal -->
    <RscsExplainer :visible="showRscsModal" @close="showRscsModal = false" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const showRscsModal = ref(false)

const { data: section, pending, error } = await useFetch(`/api/sections/${route.params.id}`)

if (error.value) {
  console.error('Error loading section:', error.value)
}

const truncatedText = computed(() => {
  if (!section.value?.text) return 'No text available.'
  const text = section.value.text
  if (text.length <= 2000) return text
  return text.substring(0, 2000) + '...'
})

function getRscsClass(rscs) {
  if (!rscs) return ''
  if (rscs < 1050) return 'rscs-low'
  if (rscs < 1150) return 'rscs-medium'
  return 'rscs-high'
}
</script>

<style scoped>
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

.rscs-low {
  color: #4d8055;
}

.rscs-medium {
  color: #c05600;
}

.rscs-high {
  color: #b50909;
}

.metric-tile {
  background-color: #f0f0f0;
  padding: 1.5rem;
  border-radius: 4px;
  text-align: center;
  border-top: 4px solid #005ea2;
}

.metric-tile__value {
  display: block;
  font-size: 2rem;
  font-weight: 700;
  color: #1b1b1b;
  line-height: 1.2;
}

.metric-tile__label {
  display: block;
  font-size: 0.875rem;
  color: #5c5c5c;
  margin-top: 0.5rem;
}

.metric-tile__weight {
  display: block;
  font-size: 0.75rem;
  color: #71767a;
  margin-top: 0.25rem;
}

.rscs-calculation {
  background-color: #e7f6f8;
  padding: 1rem;
  border-radius: 4px;
  border-left: 4px solid #00bde3;
}

.summary-content {
  background-color: #f9f9f9;
  padding: 1.5rem;
  border-left: 4px solid #005ea2;
}

.summary-content p {
  margin: 0;
  line-height: 1.6;
}

.section-text {
  background-color: #fafafa;
  padding: 1.5rem;
  border: 1px solid #dfe1e2;
  border-radius: 4px;
  font-family: 'Source Sans Pro', sans-serif;
  line-height: 1.7;
  max-height: 400px;
  overflow-y: auto;
}

.section-text p {
  margin: 0;
  white-space: pre-wrap;
}
</style>

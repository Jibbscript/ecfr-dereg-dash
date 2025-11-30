<template>
  <div class="grid-container">
    <!-- Breadcrumbs -->
    <UsaBreadcrumb class="margin-top-2">
      <UsaBreadcrumbItem href="/">Dashboard</UsaBreadcrumbItem>
      <UsaBreadcrumbItem :current="true">Title {{ route.params.t }}</UsaBreadcrumbItem>
    </UsaBreadcrumb>

    <!-- Loading State -->
    <LoadingSkeleton v-if="pending" variant="text" :count="5" />

    <!-- Error State -->
    <UsaAlert v-else-if="error" status="error" class="margin-top-2">
      <template #heading>Error Loading Title</template>
      Unable to load data for Title {{ route.params.t }}. Please try again later.
    </UsaAlert>

    <!-- Content -->
    <template v-else-if="title">
      <!-- Page Header -->
      <h1 class="font-heading-2xl margin-top-3 margin-bottom-1">
        Title {{ route.params.t }}: {{ title.title || 'Federal Regulations' }}
      </h1>

      <!-- Summary Box -->
      <UsaSummaryBox class="margin-bottom-4">
        <template #heading>Title Overview</template>
        <div class="grid-row grid-gap">
          <div class="tablet:grid-col-4">
            <p class="margin-0"><strong>Total Words:</strong></p>
            <p class="font-heading-xl margin-top-05 margin-bottom-0">
              {{ (title.total_words || 0).toLocaleString() }}
            </p>
          </div>
          <div class="tablet:grid-col-4">
            <p class="margin-0">
              <strong>Avg. RSCS Score:</strong>
              <button
                type="button"
                class="usa-button--unstyled info-icon"
                aria-label="Information about RSCS metric"
                @click="showRscsModal = true"
              >
                ⓘ
              </button>
            </p>
            <p class="font-heading-xl margin-top-05 margin-bottom-0" :class="getRscsClass(title.avg_rscs)">
              {{ (title.avg_rscs || 0).toFixed(1) }}
            </p>
          </div>
          <div class="tablet:grid-col-4">
            <p class="margin-0"><strong>Per 1,000 words</strong></p>
            <p class="text-base margin-top-05 margin-bottom-0">
              Normalized complexity score
            </p>
          </div>
        </div>
      </UsaSummaryBox>

      <!-- Accordion Sections -->
      <UsaAccordion bordered>
        <UsaAccordionItem title="Regulatory Metrics" :expanded="true">
          <div class="grid-row grid-gap">
            <div class="tablet:grid-col-6">
              <h4 class="margin-top-0">Complexity Indicators</h4>
              <ul class="usa-list">
                <li>
                  <strong>Word Count:</strong> 
                  {{ (title.total_words || 0).toLocaleString() }} total words
                </li>
                <li>
                  <strong>RSCS Score:</strong> 
                  {{ (title.avg_rscs || 0).toFixed(1) }} per 1,000 words
                </li>
              </ul>
            </div>
            <div class="tablet:grid-col-6">
              <h4 class="margin-top-0">What This Means</h4>
              <p class="text-base">
                The RSCS (Regulatory Simplicity Complexity Score) measures regulatory burden 
                by analyzing word count, definitions, cross-references, and mandatory language.
              </p>
            </div>
          </div>
        </UsaAccordionItem>

        <UsaAccordionItem title="Legislative Activity (LSA)">
          <p class="text-base">
            The List of CFR Sections Affected (LSA) tracks proposed and final rules 
            that may change regulations in this title.
          </p>
          <div class="grid-row grid-gap margin-top-2">
            <div class="tablet:grid-col-4">
              <div class="lsa-stat">
                <span class="lsa-stat__label">Proposals</span>
                <span class="lsa-stat__value">{{ title.proposals || 0 }}</span>
              </div>
            </div>
            <div class="tablet:grid-col-4">
              <div class="lsa-stat">
                <span class="lsa-stat__label">Amendments</span>
                <span class="lsa-stat__value">{{ title.amendments || 0 }}</span>
              </div>
            </div>
            <div class="tablet:grid-col-4">
              <div class="lsa-stat">
                <span class="lsa-stat__label">Final Rules</span>
                <span class="lsa-stat__value">{{ title.finals || 0 }}</span>
              </div>
            </div>
          </div>
        </UsaAccordionItem>

        <UsaAccordionItem title="AI-Generated Summary">
          <div v-if="title.summary" class="summary-content">
            <p>{{ title.summary }}</p>
          </div>
          <div v-else class="text-base text-italic">
            <p>No AI summary available for this title yet.</p>
          </div>
        </UsaAccordionItem>

        <UsaAccordionItem title="External Resources">
          <ul class="usa-list">
            <li>
              <a 
                :href="`https://www.ecfr.gov/current/title-${route.params.t}`" 
                target="_blank" 
                rel="noopener"
                class="usa-link usa-link--external"
              >
                View Title {{ route.params.t }} on eCFR.gov
              </a>
            </li>
            <li>
              <a 
                href="https://www.federalregister.gov" 
                target="_blank" 
                rel="noopener"
                class="usa-link usa-link--external"
              >
                Federal Register
              </a>
            </li>
          </ul>
        </UsaAccordionItem>
      </UsaAccordion>

      <!-- Back Link -->
      <div class="margin-top-4">
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
import { ref } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const showRscsModal = ref(false)

const { data: title, pending, error } = await useFetch(`/api/titles/${route.params.t}`)

if (error.value) {
  console.error('Error loading title:', error.value)
}

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

.lsa-stat {
  background-color: #f0f0f0;
  padding: 1rem;
  border-radius: 4px;
  text-align: center;
}

.lsa-stat__label {
  display: block;
  font-size: 0.875rem;
  color: #5c5c5c;
  margin-bottom: 0.25rem;
}

.lsa-stat__value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: #1b1b1b;
}

.summary-content {
  background-color: #f9f9f9;
  padding: 1rem;
  border-left: 4px solid #005ea2;
}

.text-italic {
  font-style: italic;
}
</style>

<template>
  <UsaModal
    v-model:visible="localVisible"
    heading="AI-Generated CFR Summaries"
    class="ai-summaries-modal"
    size="lg"
  >
    <div class="usa-prose">
      <!-- Informational Header Section -->
      <div class="info-banner">
        <div class="info-banner__icon">
          <svg aria-hidden="true" focusable="false" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
          </svg>
        </div>
        <div class="info-banner__content">
          <h3 class="info-banner__title">Comprehensive Regulatory Summaries</h3>
          <p class="info-banner__text">
            Browse AI-generated summaries for all 50 Titles of the Code of Federal Regulations.
            Each summary distills the key regulatory themes, scope, and structure of federal rules
            to help you quickly understand what each title covers.
          </p>
        </div>
      </div>

      <!-- AI Attribution Notice -->
      <div class="ai-notice">
        <div class="ai-notice__badge">
          <svg aria-hidden="true" focusable="false" width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
            <path d="M21 10.12h-6.78l2.74-2.82c-2.73-2.7-7.15-2.8-9.88-.1-2.73 2.71-2.73 7.08 0 9.79 2.73 2.71 7.15 2.71 9.88 0C18.32 15.65 19 14.08 19 12.1h2c0 1.98-.88 4.55-2.64 6.29-3.51 3.48-9.21 3.48-12.72 0-3.5-3.47-3.53-9.11-.02-12.58 3.51-3.47 9.14-3.47 12.65 0L21 3v7.12zM12.5 8v4.25l3.5 2.08-.72 1.21L11 13V8h1.5z"/>
          </svg>
          Powered by AI
        </div>
        <span class="ai-notice__model">Gemini 2.5 Pro on Vertex AI</span>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <p class="loading-text">Loading summaries...</p>
      </div>

      <!-- Error State -->
      <UsaAlert v-else-if="error" status="error" class="margin-bottom-2">
        <template #heading>Unable to Load Summaries</template>
        {{ error }}
      </UsaAlert>

      <!-- Summaries Accordion -->
      <div v-else class="usa-accordion usa-accordion--bordered summaries-accordion">
        <div v-for="title in titles" :key="title.id" class="accordion-item">
          <h4 class="usa-accordion__heading">
            <button
              class="usa-accordion__button"
              :aria-expanded="expanded === title.id"
              :aria-controls="`summary-${title.id}`"
              @click="toggle(title.id)"
            >
              <span class="accordion-title">
                <span class="title-number">Title {{ title.id }}</span>
                <span class="title-name">{{ title.name }}</span>
              </span>
              <svg class="accordion-chevron" :class="{ 'is-expanded': expanded === title.id }" aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <path d="M16.59 8.59L12 13.17 7.41 8.59 6 10l6 6 6-6z"/>
              </svg>
            </button>
          </h4>
          <div
            :id="`summary-${title.id}`"
            class="usa-accordion__content"
            :hidden="expanded !== title.id"
          >
            <div class="summary-content">
              <div class="summary-header">
                <svg class="summary-icon" aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
                </svg>
                <h5 class="summary-label">Summary Overview</h5>
              </div>
              <div class="summary-text markdown-content" v-html="renderMarkdown(title.summary)"></div>
              <div class="summary-footer">
                <span class="summary-meta">
                  <svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm.5-13H11v6l5.25 3.15.75-1.23-4.5-2.67z"/>
                  </svg>
                  Generated {{ title.date }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="!loading && !error && titles.length === 0" class="empty-state">
        <svg aria-hidden="true" width="48" height="48" viewBox="0 0 24 24" fill="currentColor">
          <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/>
        </svg>
        <p>No summaries available yet.</p>
      </div>
    </div>

    <template #footer>
      <div class="modal-footer">
        <p class="footer-note">
          <svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
          </svg>
          AI-generated content may contain inaccuracies. Verify with official sources.
        </p>
        <UsaButton @click="handleClose" variant="outline">Close</UsaButton>
      </div>
    </template>
  </UsaModal>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

function renderMarkdown(text) {
  if (!text) return ''
  const html = marked.parse(text)
  return DOMPurify.sanitize(html)
}

const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:visible'])

const localVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

function handleClose() {
  localVisible.value = false
}

const expanded = ref(null)
const titles = ref([])
const loading = ref(false)
const error = ref(null)

function toggle(id) {
  expanded.value = expanded.value === id ? null : id
}

async function fetchSummaries() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch('/api/summaries')
    if (!res.ok) throw new Error('Failed to fetch summaries')
    const data = await res.json()
    
    titles.value = data.map(s => ({
      id: s.key,
      name: getTitleName(s.key),
      summary: s.text,
      date: s.created_at ? new Date(s.created_at).toLocaleDateString() : 'Dec 1, 2025'
    })).sort((a, b) => {
      const aId = parseInt(a.id) || 0
      const bId = parseInt(b.id) || 0
      return aId - bId
    })
    
  } catch (e) {
    console.error(e)
    error.value = "Failed to load summaries. Please try again later."
  } finally {
    loading.value = false
  }
}

watch(localVisible, (val) => {
  if (val && titles.value.length === 0) {
    fetchSummaries()
  }
})

function getTitleName(id) {
  const names = {
    1: "General Provisions",
    2: "Grants and Agreements",
    3: "The President",
    4: "Accounts",
    5: "Administrative Personnel",
    6: "Domestic Security",
    7: "Agriculture",
    8: "Aliens and Nationality",
    9: "Animals and Animal Products",
    10: "Energy",
    11: "Federal Elections",
    12: "Banks and Banking",
    13: "Business Credit and Assistance",
    14: "Aeronautics and Space",
    15: "Commerce and Foreign Trade",
    16: "Commercial Practices",
    17: "Commodity and Securities Exchanges",
    18: "Conservation of Power and Water Resources",
    19: "Customs Duties",
    20: "Employees' Benefits",
    21: "Food and Drugs",
    22: "Foreign Relations",
    23: "Highways",
    24: "Housing and Urban Development",
    25: "Indians",
    26: "Internal Revenue",
    27: "Alcohol, Tobacco Products and Firearms",
    28: "Judicial Administration",
    29: "Labor",
    30: "Mineral Resources",
    31: "Money and Finance: Treasury",
    32: "National Defense",
    33: "Navigation and Navigable Waters",
    34: "Education",
    35: "Panama Canal",
    36: "Parks, Forests, and Public Property",
    37: "Patents, Trademarks, and Copyrights",
    38: "Pensions, Bonuses, and Veterans' Relief",
    39: "Postal Service",
    40: "Protection of Environment",
    41: "Public Contracts and Property Management",
    42: "Public Health",
    43: "Public Lands: Interior",
    44: "Emergency Management and Assistance",
    45: "Public Welfare",
    46: "Shipping",
    47: "Telecommunication",
    48: "Federal Acquisition Regulations System",
    49: "Transportation",
    50: "Wildlife and Fisheries"
  }
  // Handle string/number mismatch
  return names[parseInt(id)] || `Title ${id}`
}
</script>

<style scoped>
/* Info Banner */
.info-banner {
  display: flex;
  gap: 1rem;
  padding: 1.25rem;
  background: linear-gradient(135deg, #e7f2f8 0%, #d9e8f6 100%);
  border-left: 4px solid #005ea2;
  border-radius: 0 8px 8px 0;
  margin-bottom: 1.5rem;
}

.info-banner__icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  background-color: #005ea2;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.info-banner__content {
  flex: 1;
}

.info-banner__title {
  margin: 0 0 0.5rem 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: #1a4480;
}

.info-banner__text {
  margin: 0;
  font-size: 0.95rem;
  color: #3d4551;
  line-height: 1.5;
}

/* AI Notice */
.ai-notice {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background-color: #faf3d1;
  border: 1px solid #e5c000;
  border-radius: 6px;
  margin-bottom: 1.5rem;
}

.ai-notice__badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  background-color: #1a4480;
  color: white;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.ai-notice__model {
  font-size: 0.875rem;
  color: #5c5c5c;
  font-weight: 500;
}

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  gap: 1rem;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #dfe1e2;
  border-top-color: #005ea2;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  margin: 0;
  color: #5c5c5c;
  font-weight: 500;
}

/* Accordion Styling */
.summaries-accordion {
  max-height: 400px;
  overflow-y: auto;
  border-radius: 8px;
}

.accordion-item {
  border-bottom: 1px solid #dfe1e2;
}

.accordion-item:last-child {
  border-bottom: none;
}

.usa-accordion__button {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 1rem 1.25rem;
  background-color: #fff;
  border: none;
  cursor: pointer;
  text-align: left;
  transition: background-color 0.15s ease;
}

.usa-accordion__button:hover {
  background-color: #f0f0f0;
}

.accordion-title {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.title-number {
  font-size: 0.75rem;
  font-weight: 600;
  color: #005ea2;
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.title-name {
  font-size: 1rem;
  font-weight: 600;
  color: #1b1b1b;
}

.accordion-chevron {
  flex-shrink: 0;
  color: #71767a;
  transition: transform 0.2s ease;
}

.accordion-chevron.is-expanded {
  transform: rotate(180deg);
}

/* Summary Content */
.summary-content {
  background-color: #fafafa;
  padding: 1.5rem;
  border-top: 1px solid #e8e8e8;
}

.summary-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.summary-icon {
  color: #005ea2;
}

.summary-label {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: #3d4551;
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.summary-text {
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.7;
  color: #3d4551;
}

/* Markdown Content Styling */
.markdown-content :deep(h3) {
  font-size: 1.1rem;
  font-weight: 700;
  margin: 1.25rem 0 0.5rem 0;
  color: #1a4480;
  border-bottom: 1px solid #dfe1e2;
  padding-bottom: 0.375rem;
}

.markdown-content :deep(h4) {
  font-size: 1rem;
  font-weight: 600;
  margin: 1rem 0 0.5rem 0;
  color: #3d4551;
}

.markdown-content :deep(p) {
  margin: 0.5rem 0;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 0.5rem 0;
  padding-left: 1.5rem;
}

.markdown-content :deep(li) {
  margin: 0.25rem 0;
}

.markdown-content :deep(strong) {
  font-weight: 600;
  color: #1b1b1b;
}

.markdown-content :deep(em) {
  font-style: italic;
}

.markdown-content :deep(code) {
  background-color: #f0f0f0;
  padding: 0.125rem 0.375rem;
  border-radius: 3px;
  font-size: 0.875em;
}

.markdown-content :deep(blockquote) {
  border-left: 3px solid #005ea2;
  margin: 0.75rem 0;
  padding-left: 1rem;
  color: #5c5c5c;
  font-style: italic;
}

.summary-footer {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid #e8e8e8;
}

.summary-meta {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8rem;
  color: #71767a;
}

.summary-meta svg {
  color: #71767a;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: #71767a;
  text-align: center;
}

.empty-state svg {
  opacity: 0.5;
  margin-bottom: 1rem;
}

.empty-state p {
  margin: 0;
  font-size: 1rem;
}

/* Modal Footer */
.modal-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  width: 100%;
}

.footer-note {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  margin: 0;
  font-size: 0.8rem;
  color: #71767a;
}

.footer-note svg {
  flex-shrink: 0;
  color: #ffbe2e;
}

/* Scrollbar Styling */
.summaries-accordion::-webkit-scrollbar {
  width: 8px;
}

.summaries-accordion::-webkit-scrollbar-track {
  background: #f0f0f0;
  border-radius: 4px;
}

.summaries-accordion::-webkit-scrollbar-thumb {
  background: #c9c9c9;
  border-radius: 4px;
}

.summaries-accordion::-webkit-scrollbar-thumb:hover {
  background: #a9a9a9;
}
</style>

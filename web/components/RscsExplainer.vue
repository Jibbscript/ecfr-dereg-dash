<template>
  <UsaModal
    v-model:visible="localVisible"
    heading="Understanding the RSCS Metric"
    aria-describedby="rscs-modal-description"
  >
    <div id="rscs-modal-description">
      <p class="usa-intro">
        The <strong>Regulatory Simplicity Complexity Score (RSCS)</strong> is a quantitative measure 
        designed to estimate the regulatory burden and complexity of federal regulations.
      </p>

      <h3>The Formula</h3>
      <div class="formula-box">
        <code>RSCS = Words + (20 × Definitions) + (50 × Cross-refs) + (100 × Modals)</code>
      </div>
      <p class="formula-note">
        The <strong>RSCS per 1,000 words</strong> normalizes this score for comparison across 
        sections of different lengths.
      </p>

      <h3>Components Explained</h3>
      <div class="component-list">
        <div class="component-item">
          <div class="component-icon icon-blue">
            <svg aria-hidden="true" focusable="false" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
            </svg>
          </div>
          <div class="component-content">
            <h4 class="component-title">Word Count (×1)</h4>
            <p>
              The base measure of text volume. More words generally mean more to read and understand.
            </p>
          </div>
        </div>
        <div class="component-item">
          <div class="component-icon icon-green">
            <svg aria-hidden="true" focusable="false" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path d="M18 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zM6 4h5v8l-2.5-1.5L6 12V4z"/>
            </svg>
          </div>
          <div class="component-content">
            <h4 class="component-title">Definition Count (×20)</h4>
            <p>
              Terms defined using patterns like "<em>[term] means</em>" or section headers like
              "Definitions" or "As used in this part." More definitions indicate specialized
              vocabulary that requires extra cognitive effort.
            </p>
          </div>
        </div>
        <div class="component-item">
          <div class="component-icon icon-orange">
            <svg aria-hidden="true" focusable="false" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
            </svg>
          </div>
          <div class="component-content">
            <h4 class="component-title">Cross-Reference Count (×50)</h4>
            <p>
              Citations to other CFR sections (e.g., "§ 1.23" or "40 CFR 122.4"). Each cross-reference
              requires the reader to consult additional regulatory text, significantly increasing complexity.
            </p>
          </div>
        </div>
        <div class="component-item">
          <div class="component-icon icon-red">
            <svg aria-hidden="true" focusable="false" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z"/>
            </svg>
          </div>
          <div class="component-content">
            <h4 class="component-title">Modal Verb Count (×100)</h4>
            <p>
              Binding language including: <strong>shall</strong>, <strong>must</strong>,
              <strong>may not</strong>, and <strong>must not</strong>. These words create
              legal obligations and carry the highest weight because they represent actual
              regulatory requirements.
            </p>
          </div>
        </div>
      </div>

      <h3>Interpreting the Score</h3>
      <div class="interpretation-grid">
        <div class="interpretation-item low">
          <span class="score-label">Low RSCS</span>
          <span class="score-desc">Simpler, more accessible regulations</span>
        </div>
        <div class="interpretation-item high">
          <span class="score-label">High RSCS</span>
          <span class="score-desc">More complex, potentially burdensome</span>
        </div>
      </div>

      <h3>Why It Matters</h3>
      <p>
        The RSCS metric helps identify regulations that may be candidates for simplification. 
        By quantifying complexity, policymakers and the public can:
      </p>
      <ul class="usa-list">
        <li>Compare regulatory burden across agencies and titles</li>
        <li>Track changes in complexity over time</li>
        <li>Prioritize deregulation efforts where they'll have the most impact</li>
        <li>Promote transparency in the regulatory process</li>
      </ul>
    </div>

    <template #footer>
      <UsaButton @click="handleClose" class="usa-button--big width-full tablet:width-auto">
        Return to Dashboard
      </UsaButton>
    </template>
  </UsaModal>
</template>

<script setup>
import { computed } from 'vue'

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
</script>

<style scoped>
.formula-box {
  background-color: #f0f0f0;
  border-left: 4px solid #005ea2;
  padding: 1rem 1.5rem;
  margin: 1rem 0;
  font-size: 1.1rem;
  overflow-x: auto;
}

.formula-box code {
  white-space: nowrap;
}

.formula-note {
  font-size: 0.9rem;
  color: #5c5c5c;
  font-style: italic;
}

.component-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.component-item {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  background-color: #fafafa;
  border-radius: 4px;
}

.component-icon {
  width: 2.75rem;
  height: 2.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 8px;
}

.component-icon.icon-blue {
  background-color: #e7f2f8;
  color: #005ea2;
}

.component-icon.icon-green {
  background-color: #ecf3ec;
  color: #4d8055;
}

.component-icon.icon-orange {
  background-color: #fef0e1;
  color: #c05600;
}

.component-icon.icon-red {
  background-color: #f8eff0;
  color: #b50909;
}

.component-content {
  flex: 1;
}

.component-title {
  font-weight: 600;
  margin: 0 0 0.25rem 0;
  font-size: 1rem;
}

.component-content p {
  margin: 0;
  font-size: 0.9rem;
  color: #5c5c5c;
}

.interpretation-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin: 1rem 0;
}

.interpretation-item {
  padding: 1rem;
  border-radius: 4px;
  text-align: center;
}

.interpretation-item.low {
  background-color: #ecf3ec;
  border: 1px solid #4d8055;
}

.interpretation-item.high {
  background-color: #f8eff0;
  border: 1px solid #b50909;
}

.score-label {
  display: block;
  font-weight: 700;
  font-size: 1.1rem;
  margin-bottom: 0.25rem;
}

.interpretation-item.low .score-label {
  color: #4d8055;
}

.interpretation-item.high .score-label {
  color: #b50909;
}

.score-desc {
  font-size: 0.9rem;
  color: #5c5c5c;
}

h3 {
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
  color: #1b1b1b;
}

.usa-intro {
  font-size: 1.1rem;
  line-height: 1.6;
}
</style>

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
          <div class="component-icon text-blue">📝</div>
          <div class="component-content">
            <h4 class="component-title">Word Count (×1)</h4>
            <p>
              The base measure of text volume. More words generally mean more to read and understand.
            </p>
          </div>
        </div>
        <div class="component-item">
          <div class="component-icon text-green">📖</div>
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
          <div class="component-icon text-orange">🔗</div>
          <div class="component-content">
            <h4 class="component-title">Cross-Reference Count (×50)</h4>
            <p>
              Citations to other CFR sections (e.g., "§ 1.23" or "40 CFR 122.4"). Each cross-reference 
              requires the reader to consult additional regulatory text, significantly increasing complexity.
            </p>
          </div>
        </div>
        <div class="component-item">
          <div class="component-icon text-red">⚖️</div>
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
      <UsaButton @click="handleClose">Close</UsaButton>
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
  font-size: 1.5rem;
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
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

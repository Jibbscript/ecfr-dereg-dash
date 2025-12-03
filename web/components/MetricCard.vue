<template>
  <div class="usa-card metric-card">
    <div class="usa-card__container">
      <div class="usa-card__header">
        <h3 class="usa-card__heading">
          {{ title }}
          <button
            v-if="hasInfo"
            type="button"
            class="usa-button--unstyled info-button"
            :aria-label="`More information about ${title}`"
            @click="$emit('info-click', $event)"
          >
            <svg class="info-icon-svg" aria-hidden="true" focusable="false" width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/>
            </svg>
          </button>
        </h3>
      </div>
      <div class="usa-card__body">
        <p class="metric-value" :class="valueClass">
          {{ formattedValue }}
        </p>
        <p v-if="description" class="metric-description">
          {{ description }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  value: {
    type: [Number, String],
    required: true,
  },
  description: {
    type: String,
    default: '',
  },
  trend: {
    type: String,
    default: 'neutral',
    validator: (v) => ['up', 'down', 'neutral'].includes(v),
  },
  format: {
    type: String,
    default: 'number',
    validator: (v) => ['number', 'decimal', 'raw'].includes(v),
  },
  hasInfo: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['info-click'])

const formattedValue = computed(() => {
  if (props.format === 'raw') {
    return props.value
  }
  if (typeof props.value === 'number') {
    if (props.format === 'decimal') {
      return props.value.toLocaleString(undefined, {
        minimumFractionDigits: 1,
        maximumFractionDigits: 1,
      })
    }
    return props.value.toLocaleString()
  }
  return props.value
})

const valueClass = computed(() => ({
  'text-success': props.trend === 'down', // Lower RSCS is better
  'text-error': props.trend === 'up',
}))
</script>

<style scoped>
.metric-card {
  height: 100%;
}

.metric-card .usa-card__container {
  border-left: 4px solid #005ea2;
  border-radius: 0 8px 8px 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}

.metric-card .usa-card__container:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.metric-card .usa-card__header {
  padding-bottom: 0.5rem;
}

.metric-card .usa-card__heading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: #5c5c5c;
  margin-bottom: 0;
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.info-button {
  padding: 0.125rem;
  background: none;
  border: none;
  cursor: pointer;
  color: #005ea2;
  display: inline-flex;
  align-items: center;
  border-radius: 50%;
  transition: color 0.15s ease, background-color 0.15s ease;
}

.info-button:hover {
  color: #1a4480;
  background-color: rgba(0, 94, 162, 0.1);
}

.info-icon-svg {
  width: 1.125rem;
  height: 1.125rem;
}

.metric-value {
  font-size: 2.5rem;
  font-weight: 700;
  color: #1b1b1b;
  margin: 0.5rem 0;
  line-height: 1.2;
}

.metric-description {
  font-size: 0.875rem;
  color: #71767a;
  margin: 0;
}

.text-success {
  color: #4d8055;
}

.text-error {
  color: #b50909;
}
</style>

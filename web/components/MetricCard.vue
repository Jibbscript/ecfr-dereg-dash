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
            @click="$emit('info-click')"
          >
            <span class="info-icon-text" aria-hidden="true">ⓘ</span>
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
}

.metric-card .usa-card__heading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1rem;
  color: #5c5c5c;
  margin-bottom: 0;
}

.info-button {
  padding: 0;
  background: none;
  border: none;
  cursor: pointer;
  color: #005ea2;
  display: inline-flex;
  align-items: center;
}

.info-button:hover {
  color: #1a4480;
}

.info-icon-text {
  font-size: 1rem;
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

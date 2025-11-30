<template>
  <div class="skeleton-wrapper" :class="`skeleton-${variant}`" role="status" aria-label="Loading content">
    <span class="usa-sr-only">Loading...</span>
    
    <!-- Card Skeleton -->
    <template v-if="variant === 'card'">
      <div class="skeleton-card" v-for="n in count" :key="n">
        <div class="skeleton-line skeleton-title"></div>
        <div class="skeleton-line skeleton-value"></div>
        <div class="skeleton-line skeleton-desc"></div>
      </div>
    </template>

    <!-- Table Skeleton -->
    <template v-else-if="variant === 'table'">
      <div class="skeleton-table">
        <div class="skeleton-table-header">
          <div class="skeleton-line" v-for="col in columns" :key="col"></div>
        </div>
        <div class="skeleton-table-row" v-for="row in count" :key="row">
          <div class="skeleton-line" v-for="col in columns" :key="col"></div>
        </div>
      </div>
    </template>

    <!-- Text Skeleton -->
    <template v-else-if="variant === 'text'">
      <div class="skeleton-line skeleton-text" v-for="n in count" :key="n"></div>
    </template>

    <!-- Hero Skeleton -->
    <template v-else-if="variant === 'hero'">
      <div class="skeleton-hero">
        <div class="skeleton-line skeleton-hero-title"></div>
        <div class="skeleton-line skeleton-hero-subtitle"></div>
      </div>
    </template>
  </div>
</template>

<script setup>
defineProps({
  variant: {
    type: String,
    default: 'text',
    validator: (v) => ['card', 'table', 'text', 'hero'].includes(v),
  },
  count: {
    type: Number,
    default: 3,
  },
  columns: {
    type: Number,
    default: 4,
  },
})
</script>

<style scoped>
.skeleton-wrapper {
  width: 100%;
}

.skeleton-line {
  background: linear-gradient(
    90deg,
    #e0e0e0 25%,
    #f0f0f0 50%,
    #e0e0e0 75%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

/* Card Skeleton */
.skeleton-card {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

.skeleton-card {
  padding: 1.5rem;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  margin-bottom: 1rem;
}

.skeleton-title {
  height: 1rem;
  width: 60%;
  margin-bottom: 1rem;
}

.skeleton-value {
  height: 2.5rem;
  width: 40%;
  margin-bottom: 0.75rem;
}

.skeleton-desc {
  height: 0.875rem;
  width: 80%;
}

/* Table Skeleton */
.skeleton-table {
  width: 100%;
}

.skeleton-table-header {
  display: grid;
  grid-template-columns: repeat(var(--columns, 4), 1fr);
  gap: 1rem;
  padding: 1rem;
  background-color: #f0f0f0;
  margin-bottom: 0.5rem;
}

.skeleton-table-header .skeleton-line {
  height: 1rem;
}

.skeleton-table-row {
  display: grid;
  grid-template-columns: repeat(var(--columns, 4), 1fr);
  gap: 1rem;
  padding: 1rem;
  border-bottom: 1px solid #e0e0e0;
}

.skeleton-table-row .skeleton-line {
  height: 1rem;
}

/* Text Skeleton */
.skeleton-text {
  height: 1rem;
  margin-bottom: 0.75rem;
}

.skeleton-text:nth-child(odd) {
  width: 100%;
}

.skeleton-text:nth-child(even) {
  width: 85%;
}

.skeleton-text:last-child {
  width: 60%;
}

/* Hero Skeleton */
.skeleton-hero {
  padding: 3rem 0;
  text-align: center;
}

.skeleton-hero-title {
  height: 2.5rem;
  width: 50%;
  margin: 0 auto 1rem;
}

.skeleton-hero-subtitle {
  height: 1.25rem;
  width: 70%;
  margin: 0 auto;
}
</style>

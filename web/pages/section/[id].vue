<template>
  <div>
    <UsaHeading>Section {{ section.section }}</UsaHeading>
    <p>Excerpt: {{ section.text ? section.text.substring(0, 500) : '' }}...</p>
    <p>RSCS: {{ section.rscs_per_1k }}</p>
    <p>Summary: {{ section.summary }}</p>
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'

const route = useRoute()
const { data: section, error } = await useFetch(`/api/sections/${route.params.id}`)

if (error.value) {
  console.error('Error loading section:', error.value)
}
</script>

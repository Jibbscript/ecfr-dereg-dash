<template>
  <div>
    <UsaHeading>Section {{ section.section }}</UsaHeading>
    <p>Excerpt: {{ section.text ? section.text.substring(0, 500) : '' }}...</p>
    <p>RSCS: {{ section.rscs_per_1k }}</p>
    <p>Summary: {{ section.summary }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const section = ref({})

const fetchSection = async () => {
  const res = await fetch(`/api/sections/${route.params.id}`)
  section.value = await res.json()
}

fetchSection()
</script>

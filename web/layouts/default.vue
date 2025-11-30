<template>
  <div class="usa-layout">
    <!-- Skip Navigation for Accessibility -->
    <UsaSkipnav href="#main-content">Skip to main content</UsaSkipnav>

    <!-- Official Government Banner -->
    <UsaBanner />

    <!-- Site Header -->
    <UsaHeader>
      <template #logo>
        <div class="usa-logo" id="extended-logo">
          <em class="usa-logo__text">
            <NuxtLink to="/" title="Home" aria-label="Home">
              eCFR Deregulation Dashboard
            </NuxtLink>
          </em>
        </div>
      </template>

      <template #primary-navigation>
        <UsaNavPrimary :items="navItems" />
      </template>

      <template #secondary-navigation>
        <UsaSearch size="small" label="Search" @submit="handleSearch" />
      </template>
    </UsaHeader>

    <!-- Main Content Area -->
    <main id="main-content" class="main-content">
      <slot />
    </main>

    <!-- Site Footer -->
    <UsaFooter variant="big">
      <template #primary-content>
        <div class="grid-container">
          <div class="grid-row grid-gap">
            <div class="tablet:grid-col-4">
              <h3 class="usa-footer__primary-link">About</h3>
              <ul class="usa-list usa-list--unstyled">
                <li class="usa-footer__secondary-link">
                  <NuxtLink to="/">Dashboard</NuxtLink>
                </li>
                <li class="usa-footer__secondary-link">
                  <a href="https://www.ecfr.gov" target="_blank" rel="noopener">eCFR.gov</a>
                </li>
                <li class="usa-footer__secondary-link">
                  <a href="https://www.govinfo.gov" target="_blank" rel="noopener">GovInfo</a>
                </li>
              </ul>
            </div>
            <div class="tablet:grid-col-4">
              <h3 class="usa-footer__primary-link">Resources</h3>
              <ul class="usa-list usa-list--unstyled">
                <li class="usa-footer__secondary-link">
                  <a href="https://www.federalregister.gov" target="_blank" rel="noopener">Federal Register</a>
                </li>
                <li class="usa-footer__secondary-link">
                  <a href="https://www.regulations.gov" target="_blank" rel="noopener">Regulations.gov</a>
                </li>
              </ul>
            </div>
            <div class="tablet:grid-col-4">
              <h3 class="usa-footer__primary-link">Contact</h3>
              <ul class="usa-list usa-list--unstyled">
                <li class="usa-footer__secondary-link">
                  <a href="https://github.com/xai/ecfr-dereg-dashboard" target="_blank" rel="noopener">GitHub Repository</a>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </template>

      <template #secondary-content>
        <UsaIdentifier
          domain="ecfr-dashboard.gov"
          :links="identifierLinks"
        >
          <template #logo>
            <span class="text-bold">eCFR</span>
          </template>
          <template #description>
            An analytics dashboard for exploring the Electronic Code of Federal Regulations
          </template>
        </UsaIdentifier>
      </template>
    </UsaFooter>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const navItems = ref([
  {
    text: 'Dashboard',
    href: '/',
  },
  {
    text: 'About RSCS',
    href: '#',
    id: 'about-rscs-trigger',
  },
])

const identifierLinks = ref([
  {
    text: 'About this site',
    href: '/',
  },
  {
    text: 'Accessibility',
    href: '#',
  },
  {
    text: 'Privacy Policy',
    href: '#',
  },
])

function handleSearch(searchTerm) {
  console.log('Search:', searchTerm)
  // Future: implement search functionality
}
</script>

<style scoped>
.usa-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.main-content {
  flex: 1;
  padding-bottom: 2rem;
}
</style>

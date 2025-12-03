<template>
  <div class="usa-layout">
    <!-- Official Government Banner removed -->

    <!-- Site Header -->
    <UsaHeader>
      <UsaLogo>
        <template #default>
          <em class="usa-logo__text">
            <NuxtLink to="/" title="Home" aria-label="Home">
              eCFR Deregulation Dashboard
            </NuxtLink>
          </em>
        </template>
      </UsaLogo>

      <UsaNav>
        <template #primary>
          <UsaNavPrimary :items="navItems" />
        </template>
        <template #secondary>
          <button
            id="about-rscs-trigger"
            type="button"
            class="usa-button usa-button--outline"
            @click.prevent="explainer.open($event && $event.currentTarget)"
            aria-haspopup="dialog"
            aria-controls="rscs-explainer"
          >
            About RSCS
          </button>
        </template>
      </UsaNav>
    </UsaHeader>

    <!-- Main Content Area -->
    <main id="main-content" class="main-content">
      <slot />
    </main>

    <!-- Global RSCS Explainer Modal -->
    <RscsExplainer id="rscs-explainer" v-model:visible="explainer.visible.value" />

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
                  <a href="https://github.com/Jibbscript/ecfr-dereg-dashboard" target="_blank" rel="noopener">GitHub Repository</a>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </template>

      <template #secondary-content>
        <div class="grid-row flex-justify-end margin-bottom-2">
          <UsaButton variant="outline" @click="scrollToTop">Return to top</UsaButton>
        </div>
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
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import RscsExplainer from '../components/RscsExplainer.vue'
import { provideRscsExplainer } from '../composables/useRscsExplainer'

const route = useRoute()
const explainer = provideRscsExplainer()

// UsaNav items - Dashboard link only; About RSCS is a separate button
const navItems = ref([
  {
    text: 'Dashboard',
    href: '/',
  },
])

// Support deep-linking with ?rscs=1
watch(
  () => route.query.rscs,
  (val) => {
    if (val === '1' || val === 'true') explainer.open()
  },
  { immediate: true }
)

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

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
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

:deep(.usa-footer__return-to-top) {
  display: none !important;
}
</style>

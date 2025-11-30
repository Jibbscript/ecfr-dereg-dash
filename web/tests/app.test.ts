import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import App from '../app.vue'

// Mock NuxtPage
const NuxtPage = { template: '<div>Page Content</div>' }

describe('App', () => {
  it('renders NuxtPage', () => {
    const wrapper = mount(App, {
        global: {
            components: { NuxtPage }
        }
    })
    expect(wrapper.text()).toContain('Page Content')
  })
})

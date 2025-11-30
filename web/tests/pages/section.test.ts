import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SectionPage from '../../pages/section/[id].vue'
import { defineComponent, Suspense, ref } from 'vue'

// Mock vue-router
vi.mock('vue-router', () => ({
  useRoute: () => ({
    params: { id: '123' }
  })
}))

// Mock useFetch
const mockData = {
    section: '123',
    title: '40',
    text: 'Sample text content for section 123...',
    rscs_per_1k: 10.5,
    summary: 'A brief summary.'
}

global.useFetch = vi.fn().mockResolvedValue({
    data: ref(mockData),
    error: ref(null)
})

describe('SectionPage', () => {
  it('renders section details', async () => {
    // Wrap in Suspense because of top-level await
    const TestComponent = defineComponent({
        components: { SectionPage },
        template: '<Suspense><SectionPage /></Suspense>'
    })

    const wrapper = mount(TestComponent)

    await flushPromises()
    
    expect(wrapper.text()).toContain('123') // heading contains section number
    expect(wrapper.text()).toContain('Sample text')
    // RSCS appears as a numeric value with per 1,000 words context
    expect(wrapper.text()).toContain('10.5')
  })
  
  it('logs error on failure', async () => {
       const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
       global.useFetch = vi.fn().mockResolvedValue({
            data: ref(null),
            error: ref('Fetch failed')
        })
        
        const TestComponent = defineComponent({
            components: { SectionPage },
            template: '<Suspense><SectionPage /></Suspense>'
        })

        mount(TestComponent)
        
        await flushPromises()
        
        expect(consoleSpy).toHaveBeenCalledWith('Error loading section:', 'Fetch failed')
        consoleSpy.mockRestore()
  })
})
